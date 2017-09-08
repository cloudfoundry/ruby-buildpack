package supply

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"ruby/cache"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/kr/text"
)

type Command interface {
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(string, string, ...string) (string, error)
	Run(*exec.Cmd) error
}

type Manifest interface {
	AllDependencyVersions(string) []string
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
	DefaultVersion(string) (libbuildpack.Dependency, error)
}
type Versions interface {
	Engine() (string, error)
	Version() (string, error)
	JrubyVersion() (string, error)
	RubyEngineVersion() (string, error)
	HasGemVersion(gem string, constraints ...string) (bool, error)
	VersionConstraint(version string, constraints ...string) (bool, error)
	HasWindowsGemfileLock() (bool, error)
	Gemfile() string
}

type Stager interface {
	BuildDir() string
	DepDir() string
	DepsIdx() string
	LinkDirectoryInDepDir(string, string) error
	WriteEnvFile(string, string) error
	WriteProfileD(string, string) error
	SetStagingEnvironment() error
}

type Cache interface {
	Metadata() *cache.Metadata
	Restore() error
	Save() error
}

type Supplier struct {
	Stager          Stager
	Manifest        Manifest
	Log             *libbuildpack.Logger
	Versions        Versions
	Cache           Cache
	Command         Command
	cachedNeedsNode bool
	needsNode       bool
}

func Run(s *Supplier) error {
	s.Log.BeginStep("Supplying Ruby")

	_ = s.Command.Execute(s.Stager.BuildDir(), ioutil.Discard, ioutil.Discard, "touch", "/tmp/checkpoint")

	if err := s.AssetGemfileLockExists(); err != nil {
		s.Log.Error(err.Error())
		return err
	}

	if checksum, err := s.CalcChecksum(); err == nil {
		s.Log.Debug("BuildDir Checksum Before Supply: %s", checksum)
	}

	if err := s.Cache.Restore(); err != nil {
		s.Log.Error("Unable to restore cache: %s", err.Error())
		return err
	}

	if err := s.InstallBundler(); err != nil {
		s.Log.Error("Unable to install bundler: %s", err.Error())
		return err
	}

	if err := s.CreateDefaultEnv(); err != nil {
		s.Log.Error("Unable to setup default environment: %s", err.Error())
		return err
	}

	engine, rubyVersion, err := s.DetermineRuby()
	if err != nil {
		s.Log.Error(err.Error())
		return err
	}

	if engine == "jruby" {
		if err = s.InstallJVM(); err != nil {
			s.Log.Error("Unable to install JVM: %s", err.Error())
			return err
		}
	}

	if err := s.InstallRuby(engine, rubyVersion); err != nil {
		s.Log.Error("Unable to install ruby: %s", err.Error())
		return err
	}

	if err := s.AddPostRubyInstallDefaultEnv(engine); err != nil {
		s.Log.Error("Unable to add bundler and gem path to default environment: %s", err.Error())
		return err
	}

	if err := s.UpdateRubygems(); err != nil {
		s.Log.Error("Unable to update rubygems: %s", err.Error())
		return err
	}

	if s.NeedsNode() {
		if err := s.InstallNode(); err != nil {
			s.Log.Error("Unable to install node: %s", err.Error())
			return err
		}

		if err := s.InstallYarn(); err != nil {
			s.Log.Error("Unable to install yarn: %s", err.Error())
			return err
		}
	}

	if err := s.InstallGems(); err != nil {
		s.Log.Error("Unable to install gems: %s", err.Error())
		return err
	}

	if err := s.WriteProfileD(engine); err != nil {
		s.Log.Error("Unable to write profile.d: %s", err.Error())
		return err
	}

	if err := s.Cache.Save(); err != nil {
		s.Log.Error("Unable to save cache: %s", err.Error())
		return err
	}

	if err := s.Stager.SetStagingEnvironment(); err != nil {
		s.Log.Error("Unable to setup environment variables: %s", err.Error())
		return err
	}

	if checksum, err := s.CalcChecksum(); err == nil {
		s.Log.Debug("BuildDir Checksum After Supply: %s", checksum)
	}

	if filesChanged, err := s.Command.Output(s.Stager.BuildDir(), "find", ".", "-newer", "/tmp/checkpoint", "-not", "-path", "./.cloudfoundry/*", "-not", "-path", "./.cloudfoundry"); err == nil && filesChanged != "" {
		s.Log.Debug("Below files changed:")
		s.Log.Debug(filesChanged)
	}

	return nil
}

func (s *Supplier) AssetGemfileLockExists() error {
	if exists, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), "Gemfile.lock")); err != nil {
		return err
	} else if !exists {
		return errors.New("Gemfile.lock required")
	}
	return nil
}

func (s *Supplier) DetermineRuby() (string, string, error) {
	engine, err := s.Versions.Engine()
	if err != nil {
		return "", "", fmt.Errorf("Unable to determine ruby engine: %v", err)
	}

	var rubyVersion string
	if engine == "ruby" {
		rubyVersion, err = s.Versions.Version()
		if err != nil {
			return "", "", fmt.Errorf("Unable to determine ruby version: %v", err)
		}
		if rubyVersion == "" {
			if dep, err := s.Manifest.DefaultVersion("ruby"); err != nil {
				return "", "", fmt.Errorf("Unable to determine ruby version: %v", err)
			} else {
				rubyVersion = dep.Version
				s.Log.Warning("You have not declared a Ruby version in your Gemfile.\nDefaulting to %s\nSee http://docs.cloudfoundry.org/buildpacks/ruby/index.html#runtime for more information.", rubyVersion)
			}
		}
	} else if engine == "jruby" {
		rubyVersion, err = s.Versions.JrubyVersion()
		if err != nil {
			return "", "", fmt.Errorf("Unable to determine jruby version: %v", err)
		}
	} else {
		return "", "", fmt.Errorf("Sorry, we do not support engine: %s", engine)
	}
	return engine, rubyVersion, nil
}

func (s *Supplier) InstallYarn() error {
	exists, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), "yarn.lock"))
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	tempDir, err := ioutil.TempDir("", "yarn")
	if err != nil {
		return err
	}
	if err := s.Manifest.InstallOnlyVersion("yarn", tempDir); err != nil {
		return err
	}
	if paths, err := filepath.Glob(filepath.Join(tempDir, "yarn-v*")); err != nil {
		return err
	} else if len(paths) != 1 {
		return fmt.Errorf("Unable to find yarn distribution dir")
	} else {
		tempDir = paths[0]
	}

	if err := os.Rename(tempDir, filepath.Join(s.Stager.DepDir(), "yarn")); err != nil {
		return err
	}
	return s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "yarn", "bin"), "bin")
}

func (s *Supplier) InstallBundler() error {
	if err := s.Manifest.InstallOnlyVersion("bundler", filepath.Join(s.Stager.DepDir(), "bundler")); err != nil {
		return err
	}

	if err := s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "bundler", "bin"), "bin"); err != nil {
		return err
	}

	return nil
}

func (s *Supplier) InstallNode() error {
	var dep libbuildpack.Dependency

	tempDir, err := ioutil.TempDir("", "node")
	if err != nil {
		return err
	}
	nodeInstallDir := filepath.Join(s.Stager.DepDir(), "node")

	version := "4.x"
	rails51, err := s.Versions.HasGemVersion("rails", ">=5.1.0.beta")
	if err != nil {
		return err
	}
	if rails51 {
		version = "6.x"
	}

	versions := s.Manifest.AllDependencyVersions("node")
	ver, err := libbuildpack.FindMatchingVersion(version, versions)
	if err != nil {
		return err
	}
	dep.Name = "node"
	dep.Version = ver

	if err := s.Manifest.InstallDependency(dep, tempDir); err != nil {
		return err
	}

	if err := os.Rename(filepath.Join(tempDir, fmt.Sprintf("node-v%s-linux-x64", dep.Version)), nodeInstallDir); err != nil {
		return err
	}

	return s.Stager.LinkDirectoryInDepDir(filepath.Join(nodeInstallDir, "bin"), "bin")
}

func (s *Supplier) NeedsNode() bool {
	if s.cachedNeedsNode {
		return s.needsNode
	}
	s.cachedNeedsNode = true
	s.needsNode = false

	if s.isNodeInstalled() {
		s.Log.BeginStep("Skipping install of nodejs since it has been supplied")
	} else {
		for _, name := range []string{"webpacker", "execjs"} {
			s.Log.Debug("Test %s in gemfile", name)
			hasgem, err := s.Versions.HasGemVersion(name, ">=0.0.0")
			if err == nil && hasgem {
				s.Log.Debug("Found %s in gemfile", name)
				s.needsNode = true
				break
			}
		}
	}

	return s.needsNode
}

func (s *Supplier) isNodeInstalled() bool {
	_, err := s.Command.Output(s.Stager.BuildDir(), "node", "--version")
	return err == nil
}

func (s *Supplier) InstallJVM() error {
	if exists, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), ".jdk")); err != nil {
		return err
	} else if exists {
		s.Log.Info("Using pre-installed JDK")
		return nil
	}

	jvmInstallDir := filepath.Join(s.Stager.DepDir(), "jvm")
	if err := s.Manifest.InstallOnlyVersion("openjdk1.8-latest", jvmInstallDir); err != nil {
		return err
	}
	if err := s.Stager.LinkDirectoryInDepDir(filepath.Join(jvmInstallDir, "bin"), "bin"); err != nil {
		return err
	}

	scriptContents := `
if ! [[ "${JAVA_OPTS}" == *-Xmx* ]]; then
  export JAVA_MEM=${JAVA_MEM:--Xmx${JVM_MAX_HEAP:-384}m}
fi
export JAVA_OPTS=${JAVA_OPTS:--Xss512k -XX:+UseCompressedOops -Dfile.encoding=UTF-8}
export JRUBY_OPTS=${JRUBY_OPTS:--Xcompile.invokedynamic=false}
`

	return s.Stager.WriteProfileD("jruby.sh", scriptContents)
}

func (s *Supplier) InstallRuby(name, version string) error {
	installDir := filepath.Join(s.Stager.DepDir(), "ruby")

	if err := s.Manifest.InstallDependency(libbuildpack.Dependency{Name: name, Version: version}, installDir); err != nil {
		return err
	}

	rakePath := filepath.Join(s.Stager.DepDir(), "ruby", "bin", "rake")
	rakeContent, err := ioutil.ReadFile(rakePath)
	if err != nil {
		return err
	}
	groups := strings.SplitN(string(rakeContent), "\n", 2)
	groups[0] = fmt.Sprintf("#!/usr/bin/env %s", name)
	if err := ioutil.WriteFile(rakePath, []byte(strings.Join(groups, "\n")), 0755); err != nil {
		return err
	}

	if err := os.Symlink("ruby", filepath.Join(s.Stager.DepDir(), "ruby", "bin", "ruby.exe")); err != nil {
		return err
	}
	return s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "ruby", "bin"), "bin")
}

func (s *Supplier) UpdateRubygems() error {
	dep := libbuildpack.Dependency{Name: "rubygems"}
	versions := s.Manifest.AllDependencyVersions(dep.Name)
	if len(versions) == 0 {
		return nil
	} else if len(versions) > 1 {
		return fmt.Errorf("Too many versions of rubygems in manifest")
	}
	dep.Version = versions[0]

	currVersion, err := s.Command.Output("/", "gem", "--version")
	if err != nil {
		return fmt.Errorf("Could not determine current version of rubygems: %v", err)
	}

	currVersion = strings.TrimSpace(currVersion)
	if newer, err := s.Versions.VersionConstraint(currVersion, fmt.Sprintf(">= %s", dep.Version)); err != nil {
		return fmt.Errorf("Could not parse rubygems version constraint: %s >= %s: %v", currVersion, dep.Version, err)
	} else if newer {
		return nil
	}

	s.Log.BeginStep("Update rubygems from %s to %s", currVersion, dep.Version)

	tempDir, err := ioutil.TempDir("", "node")
	if err != nil {
		return err
	}

	if err := s.Manifest.InstallDependency(dep, tempDir); err != nil {
		return err
	}

	rubygemsDir := filepath.Join(tempDir, fmt.Sprintf("rubygems-%s", dep.Version))
	if output, err := s.Command.Output(rubygemsDir, "ruby", "setup.rb"); err != nil {
		s.Log.Error(output)
		return fmt.Errorf("Could not install rubygems: %v", err)
	}

	return nil
}

type IndentedWriter struct {
	w   io.Writer
	pad string
}

func (w *IndentedWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for i, line := range lines {
		lines[i] = w.pad + line
	}
	p = []byte(strings.Join(lines, "\n"))
	return w.Write(p)
}

func (s *Supplier) copyDirToTemp(dir string) (string, error) {
	tempDir, err := ioutil.TempDir("", "app")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("cp", "-al", dir, tempDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		s.Log.Error(string(output))
		return "", fmt.Errorf("Could not copy build dir to temp: %v", err)
	}
	tempDir = filepath.Join(tempDir, filepath.Base(dir))
	return tempDir, nil
}

func (s *Supplier) InstallGems() error {
	s.warnBundleConfig()
	s.warnWindowsGemfile()

	tempDir, err := s.copyDirToTemp(s.Stager.BuildDir())
	if err != nil {
		return nil
	}
	gemfileLock, err := filepath.Rel(s.Stager.BuildDir(), s.Versions.Gemfile())
	if err != nil {
		return nil
	}
	gemfileLock = filepath.Join(tempDir, gemfileLock) + ".lock"

	if hasFile, err := s.Versions.HasWindowsGemfileLock(); err != nil {
		return err
	} else if hasFile {
		s.Log.Debug("Remove %s", gemfileLock)
		s.Log.Warning("Removing `Gemfile.lock` because it was generated on Windows.\nBundler will do a full resolve so native gems are handled properly.\nThis may result in unexpected gem versions being used in your app.\nIf you are using multi buildpacks, subsequent buildpacks may fail.\nIn rare occasions Bundler may not be able to resolve your dependencies at all.\nhttps://docs.cloudfoundry.org/buildpacks/ruby/windows.html")
		if err := os.Remove(gemfileLock); err != nil {
			return fmt.Errorf("Remove Gemfile.lock: %v", err)
		}
	}

	args := []string{"install", "--without", os.Getenv("BUNDLE_WITHOUT"), "--jobs=4", "--retry=4", "--path", filepath.Join(s.Stager.DepDir(), "vendor_bundle"), "--binstubs", filepath.Join(s.Stager.DepDir(), "binstubs")}
	if exists, err := libbuildpack.FileExists(gemfileLock); err != nil {
		return err
	} else if exists {
		args = append(args, "--deployment")
	}

	version := s.Manifest.AllDependencyVersions("bundler")
	s.Log.BeginStep("Installing dependencies using bundler %s", version[0])
	s.Log.Info("Running: bundle %s", strings.Join(args, " "))

	env := os.Environ()
	env = append(env, "NOKOGIRI_USE_SYSTEM_LIBRARIES=true")

	cmd := exec.Command("bundle", args...)
	cmd.Dir = tempDir
	cmd.Stdout = text.NewIndentWriter(os.Stdout, []byte("       "))
	cmd.Stderr = text.NewIndentWriter(os.Stderr, []byte("       "))
	cmd.Env = env
	if err := s.Command.Run(cmd); err != nil {
		return err
	}

	s.Log.Info("Cleaning up the bundler cache.")

	cmd = exec.Command("bundle", "clean")
	cmd.Dir = tempDir
	cmd.Stdout = text.NewIndentWriter(os.Stdout, []byte("       "))
	cmd.Stderr = text.NewIndentWriter(os.Stderr, []byte("       "))
	cmd.Env = env
	if err := s.Command.Run(cmd); err != nil {
		return err
	}

	// Copy binstubs to bin
	files, err := ioutil.ReadDir(filepath.Join(s.Stager.DepDir(), "binstubs"))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Could not read dep/binstubs directory: %v", err)
		}
	} else {
		for _, file := range files {
			source := filepath.Join(s.Stager.DepDir(), "binstubs", file.Name())
			target := filepath.Join(s.Stager.DepDir(), "bin", file.Name())
			if exists, err := libbuildpack.FileExists(target); err != nil {
				return fmt.Errorf("Checking existence: %v", err)
			} else if !exists {
				if err := libbuildpack.CopyFile(source, target); err != nil {
					return fmt.Errorf("CopyFile: %v", err)
				}
			}
		}
	}

	// Save .bundle/config to global config
	if exists, err := libbuildpack.FileExists(filepath.Join(tempDir, ".bundle", "config")); err == nil && exists {
		s.Log.Debug("SaveGemfileLock; %s -> %s", filepath.Join(tempDir, ".bundle", "config"), os.Getenv("BUNDLE_CONFIG"))
		if err := os.Rename(filepath.Join(tempDir, ".bundle", "config"), os.Getenv("BUNDLE_CONFIG")); err != nil {
			return err
		}
	}

	// Save Gemfile.lock for finalize
	gemfileLockTarget := filepath.Join(s.Stager.DepDir(), "Gemfile.lock")
	if exists, err := libbuildpack.FileExists(gemfileLock); err == nil && exists {
		s.Log.Debug("SaveGemfileLock; %s -> %s", gemfileLock, gemfileLockTarget)
		if err := os.Rename(gemfileLock, gemfileLockTarget); err != nil {
			return err
		}
	}

	return os.RemoveAll(tempDir)
}

func (s *Supplier) CreateDefaultEnv() error {
	environmentDefaults := map[string]string{
		"RAILS_ENV":      "production",
		"RACK_ENV":       "production",
		"RAILS_GROUPS":   "assets",
		"BUNDLE_WITHOUT": "development:test",
		"BUNDLE_BIN":     filepath.Join(s.Stager.DepDir(), "binstubs"),
		"BUNDLE_CONFIG":  filepath.Join(s.Stager.DepDir(), "bundle_config"),
		"GEM_HOME":       filepath.Join(s.Stager.DepDir(), "gem_home"),
		"GEM_PATH": strings.Join([]string{
			filepath.Join(s.Stager.DepDir(), "gem_home"),
			filepath.Join(s.Stager.DepDir(), "bundler"),
		}, ":"),
	}
	return s.writeEnvFiles(environmentDefaults, false)
}

func (s *Supplier) AddPostRubyInstallDefaultEnv(engine string) error {
	rubyEngineVersion, err := s.Versions.RubyEngineVersion()
	if err != nil {
		s.Log.Error("Unable to determine ruby engine: %s", err.Error())
		return err
	}
	environmentDefaults := map[string]string{
		"BUNDLE_PATH": filepath.Join(s.Stager.DepDir(), "vendor_bundle", engine, rubyEngineVersion),
		"GEM_PATH": strings.Join([]string{
			filepath.Join(s.Stager.DepDir(), "vendor_bundle", engine, rubyEngineVersion),
			filepath.Join(s.Stager.DepDir(), "gem_home"),
			filepath.Join(s.Stager.DepDir(), "bundler"),
		}, ":"),
	}
	s.Log.Debug("Setting post ruby install env: %v", environmentDefaults)
	return s.writeEnvFiles(environmentDefaults, true)
}

func (s *Supplier) writeEnvFiles(environment map[string]string, clobber bool) error {
	for envVar, envDefault := range environment {
		if os.Getenv(envVar) == "" || clobber {
			_ = os.Setenv(envVar, envDefault)
			if err := s.Stager.WriteEnvFile(envVar, envDefault); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Supplier) WriteProfileD(engine string) error {
	s.Log.BeginStep("Creating runtime environment")

	rubyEngineVersion, err := s.Versions.RubyEngineVersion()
	if err != nil {
		return err
	}

	depsIdx := s.Stager.DepsIdx()
	scriptContents := fmt.Sprintf(`
export LANG=${LANG:-en_US.UTF-8}
export RAILS_ENV=${RAILS_ENV:-production}
export RACK_ENV=${RACK_ENV:-production}
export RAILS_SERVE_STATIC_FILES=${RAILS_SERVE_STATIC_FILES:-enabled}
export RAILS_LOG_TO_STDOUT=${RAILS_LOG_TO_STDOUT:-enabled}

export GEM_HOME=${GEM_HOME:-$DEPS_DIR/%s/gem_home}
export GEM_PATH=${GEM_PATH:-$DEPS_DIR/%s/vendor_bundle/%s/%s:$DEPS_DIR/%s/gem_home:$DEPS_DIR/%s/bundler}
export BUNDLE_PATH=${BUNDLE_PATH:-$DEPS_DIR/%s/vendor_bundle/%s/%s}

## Change to current DEPS_DIR
bundle config PATH "$DEPS_DIR/%s/vendor_bundle" > /dev/null
bundle config WITHOUT "%s" > /dev/null
`, depsIdx, depsIdx, engine, rubyEngineVersion, depsIdx, depsIdx, depsIdx, engine, rubyEngineVersion, depsIdx, os.Getenv("BUNDLE_WITHOUT"))

	hasRails41, err := s.Versions.HasGemVersion("rails", ">=4.1.0.beta1")
	if err != nil {
		return err
	}
	if hasRails41 {
		metadata := s.Cache.Metadata()
		if metadata.SecretKeyBase == "" {
			metadata.SecretKeyBase, err = s.Command.Output(s.Stager.BuildDir(), "bundle", "exec", "rake", "secret")
			if err != nil {
				return fmt.Errorf("Running 'rake secret'", err)
			}
		}
		scriptContents += fmt.Sprintf("\nexport SECRET_KEY_BASE=${SECRET_KEY_BASE:-%s}\n", metadata.SecretKeyBase)
	}

	return s.Stager.WriteProfileD("ruby.sh", scriptContents)
}

func (s *Supplier) CalcChecksum() (string, error) {
	h := md5.New()
	basepath := s.Stager.BuildDir()
	err := filepath.Walk(basepath, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			relpath, err := filepath.Rel(basepath, path)
			if strings.HasPrefix(relpath, ".cloudfoundry/") {
				return nil
			}
			if err != nil {
				return err
			}
			if _, err := io.WriteString(h, relpath); err != nil {
				return err
			}
			if f, err := os.Open(path); err != nil {
				return err
			} else {
				if _, err := io.Copy(h, f); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (s *Supplier) warnWindowsGemfile() {
	if body, err := ioutil.ReadFile(filepath.Join(s.Stager.BuildDir(), "Gemfile")); err == nil {
		if bytes.Contains(body, []byte("\r\n")) {
			s.Log.Warning("Windows line endings detected in Gemfile. Your app may fail to stage. Please use UNIX line endings.")
		}
	}
}

func (s *Supplier) warnBundleConfig() {
	if exists, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), ".bundle", "config")); err == nil && exists {
		s.Log.Warning("You have the `.bundle/config` file checked into your repository\nIt contains local state like the location of the installed bundle\nas well as configured git local gems, and other settings that should\nnot be shared between multiple checkouts of a single repo. Please\nremove the `.bundle/` folder from your repo and add it to your `.gitignore` file.")
	}
}
