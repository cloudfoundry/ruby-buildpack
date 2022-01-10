package supply

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/ruby-buildpack/src/ruby/cache"
	"github.com/kr/text"
)

type Command interface {
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(string, string, ...string) (string, error)
	Run(*exec.Cmd) error
}

type Manifest interface {
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Installer interface {
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Versions interface {
	GetBundlerVersion() (string, error)
	Engine() (string, error)
	Version() (string, error)
	JrubyVersion() (string, error)
	RubyEngineVersion() (string, error)
	HasGemVersion(gem string, constraints ...string) (bool, error)
	VersionConstraint(version string, constraints ...string) (bool, error)
	HasWindowsGemfileLock() (bool, error)
	Gemfile() string
	BundledWithVersion() (string, error)
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

type TempDir interface {
	CopyDirToTemp(string) (string, error)
}

type Cache interface {
	Metadata() *cache.Metadata
	Restore() error
	Save() error
}

type Supplier struct {
	Stager            Stager
	Manifest          Manifest
	Installer         Installer
	Log               *libbuildpack.Logger
	Versions          Versions
	Cache             Cache
	Command           Command
	TempDir           TempDir
	cachedNeedsNode   bool
	needsNode         bool
	appHasGemfile     bool
	appHasGemfileLock bool
}

func Run(s *Supplier) error {
	s.Log.BeginStep("Supplying Ruby")

	_ = s.Command.Execute(s.Stager.BuildDir(), ioutil.Discard, ioutil.Discard, "touch", "/tmp/checkpoint")

	if checksum, err := s.CalcChecksum(); err == nil {
		s.Log.Debug("BuildDir Checksum Before Supply: %s", checksum)
	}

	if err := s.Setup(); err != nil {
		s.Log.Error("Error during setup: %v", err)
		return err
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

	if err := s.EnableLDLibraryPathEnv(); err != nil {
		s.Log.Error("Unable to enable ld_library_path env: %s", err.Error())
		return err
	}

	engine, rubyVersion, err := s.DetermineRuby()
	if err != nil {
		s.Log.Error("Unable to determine ruby: %s", err.Error())
		return err
	}

	if engine == "jruby" {
		if err = s.InstallJVM(); err != nil {
			s.Log.Error("Unable to install JVM: %s", err.Error())
			return err
		}
	}

	// Search cache dir for sub-directories which don't match current ruby version.
	if err := s.RemoveUnusedRubyVersions(engine, rubyVersion); err != nil {
		s.Log.Error("Unable to remove unused ruby: %s", err.Error())
		return err
	}

	if err := s.InstallRuby(engine, rubyVersion); err != nil {
		s.Log.Error("Unable to install ruby: %s", err.Error())
		return err
	}

	if err := s.AddPostRubyInstallDefaultEnv(engine); err != nil {
		s.Log.Error("Unable to add gem path to default environment: %s", err.Error())
		return err
	}

	if err := s.UpdateRubygems(); err != nil {
		s.Log.Error("Unable to update rubygems: %s", err.Error())
		return err
	}

	if err := s.AddPostRubyGemsInstallDefaultEnv(engine); err != nil {
		s.Log.Error("Unable to add bundler path to default environment: %s", err.Error())
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

	if err := s.RewriteShebangs(); err != nil {
		s.Log.Error("Unable to rewrite shebangs: %s", err.Error())
		return err
	}

	if err := s.SymlinkBundlerIntoRubygems(); err != nil {
		s.Log.Error("Unable to symlink bundler into rubygems: %s", err.Error())
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

func (s *Supplier) Setup() error {
	if exists, err := libbuildpack.FileExists(s.Versions.Gemfile()); err != nil {
		return fmt.Errorf("unable to determine if Gemfile exists: %v", err)
	} else {
		s.appHasGemfile = exists
	}

	if exists, err := libbuildpack.FileExists(fmt.Sprintf("%s.lock", s.Versions.Gemfile())); err != nil {
		return fmt.Errorf("Unable to determine if Gemfile.lock exists: %v", err)
	} else {
		s.appHasGemfileLock = exists
	}

	return nil
}

func (s *Supplier) DetermineRuby() (string, string, error) {
	if !s.appHasGemfile {
		dep, err := s.Manifest.DefaultVersion("ruby")
		if err != nil {
			return "", "", fmt.Errorf("unable to determine default ruby version: %v", err)
		}
		return "ruby", dep.Version, nil
	}

	engine, err := s.Versions.Engine()
	if err != nil {
		return "", "", fmt.Errorf("unable to determine ruby engine: %v", err)
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

func (s *Supplier) RemoveUnusedRubyVersions(engine, version string) error {
	splitVersion := strings.Split(version, ".")
	majorMinorVersion := strings.Join([]string{splitVersion[0], splitVersion[1], "0"}, ".")

	matches, err := filepath.Glob(filepath.Join(s.Stager.DepDir(), "vendor_bundle", engine, "*"))
	if err != nil {
		return err
	}

	for _, match := range matches {
		if !strings.HasSuffix(match, majorMinorVersion) {
			err := os.RemoveAll(match)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Supplier) InstallYarn() error {
	exists, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), "yarn.lock"))
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	yarnInstallDir := filepath.Join(s.Stager.DepDir(), "yarn")
	if err != nil {
		return err
	}
	if err := s.Installer.InstallOnlyVersion("yarn", yarnInstallDir); err != nil {
		return err
	}

	return s.Stager.LinkDirectoryInDepDir(filepath.Join(yarnInstallDir, "bin"), "bin")
}

func (s *Supplier) InstallBundler() error {
	contents, err := ioutil.ReadFile(fmt.Sprintf("%s.lock", s.Versions.Gemfile()))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	re := regexp.MustCompile(`BUNDLED WITH\s*(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(contents))

	if len(matches) != 2 {
		matches = []string{"", "2"}
	}

	if strings.HasPrefix(matches[1], "2") {
		return s.installBundler("2.x.x")
	}

	return s.installBundler("1.x.x")
}

func (s *Supplier) InstallNode() error {
	var dep libbuildpack.Dependency

	version, err := libbuildpack.FindMatchingVersion("x", s.Manifest.AllDependencyVersions("node"))
	if err != nil {
		return err
	}
	dep.Name = "node"
	dep.Version = version

	nodeInstallDir := filepath.Join(s.Stager.DepDir(), "node")
	if err := s.Installer.InstallDependency(dep, nodeInstallDir); err != nil {
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
		for _, name := range []string{"webpacker", "execjs", "cssbundling-rails", "jsbundling-rails"} {
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
	if err := s.Installer.InstallOnlyVersion("openjdk1.8-latest", jvmInstallDir); err != nil {
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
	installDir := filepath.Join(s.Stager.DepDir(), name)

	if err := s.Installer.InstallDependency(libbuildpack.Dependency{Name: name, Version: version}, installDir); err != nil {
		return err
	}

	if err := s.RewriteShebangs(); err != nil {
		return err
	}

	if err := os.Symlink("ruby", filepath.Join(installDir, "bin", "ruby.exe")); err != nil {
		return err
	}

	return s.Stager.LinkDirectoryInDepDir(filepath.Join(installDir, "bin"), "bin")
}

func (s *Supplier) VendorBundlePath() (string, error) {
	bundlerVersion, err := s.Versions.GetBundlerVersion()
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(bundlerVersion, "2.") {
		return "vendor_bundle", nil
	}

	engine, err := s.Versions.Engine()
	if err != nil {
		return "", fmt.Errorf("Unable to determine ruby engine: %s", err)
	}

	rubyEngineVersion, err := s.Versions.RubyEngineVersion()
	if err != nil {
		return "", fmt.Errorf("Unable to determine ruby engine version: %s", err)
	}

	return filepath.Join("vendor_bundle", engine, rubyEngineVersion), nil
}

func (s *Supplier) RewriteShebangs() error {
	engine, err := s.Versions.Engine()
	if err != nil {
		return err
	}

	files1, err := filepath.Glob(filepath.Join(s.Stager.DepDir(), "bin", "*"))
	if err != nil {
		return err
	}

	files2, err := filepath.Glob(filepath.Join(s.Stager.DepDir(), "vendor_bundle", engine, "*", "bin", "*"))
	if err != nil {
		return err
	}

	for _, file := range append(files1, files2...) {
		if fileInfo, err := os.Stat(file); err != nil {
			return err
		} else if fileInfo.IsDir() {
			continue
		}

		fileContents, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		shebangRegex := regexp.MustCompile(`^#!/.*ruby.*`)
		fileContents = shebangRegex.ReplaceAll(fileContents, []byte("#!/usr/bin/env ruby"))
		if err := ioutil.WriteFile(file, fileContents, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Supplier) SymlinkBundlerIntoRubygems() error {
	s.Log.Debug("SymlinkBundlerIntoRubygems")

	engine, err := s.Versions.Engine()
	if err != nil {
		return err
	}

	rubyEngineVersion, err := s.Versions.RubyEngineVersion()
	if err != nil {
		return fmt.Errorf("Unable to determine ruby engine: %s", err)
	}

	bundlerVersion, err := s.Versions.GetBundlerVersion()
	if err != nil {
		return err
	}

	destDir := filepath.Join(s.Stager.DepDir(), engine, "lib", "ruby", "gems", rubyEngineVersion, "gems")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	srcDir := filepath.Join(s.Stager.DepDir(), "bundler", "gems", "bundler-"+bundlerVersion)
	relPath, err := filepath.Rel(destDir, srcDir)
	if err != nil {
		return err
	}

	destFile := filepath.Join(destDir, "bundler-"+bundlerVersion)
	if found, err := libbuildpack.FileExists(destFile); err != nil {
		return err
	} else if found {
		s.Log.Debug("Skipping linking bundler since destination exists")
		return nil
	}

	return os.Symlink(relPath, destFile)
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

	if engine, err := s.Versions.Engine(); err != nil {
		return err
	} else if engine == "jruby" {
		s.Log.Debug("Skipping update of rubygems since jruby")
		return nil
	}

	s.Log.BeginStep("Update rubygems from %s to %s", currVersion, dep.Version)

	tempDir, err := ioutil.TempDir("", "rubygems")
	if err != nil {
		return err
	}

	if err := s.Installer.InstallDependency(dep, tempDir); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(s.Stager.DepDir(), "gem_home"), 0755); err != nil {
		return err
	}

	if output, err := s.Command.Output(tempDir, "ruby", "setup.rb"); err != nil {
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

type LinuxTempDir struct {
	Log *libbuildpack.Logger
}

func (t *LinuxTempDir) CopyDirToTemp(dir string) (string, error) {
	tempDir, err := ioutil.TempDir("", "app")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("cp", "-al", dir, tempDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Log.Error(string(output))
		return "", fmt.Errorf("Could not copy build dir to temp: %v", err)
	}
	tempDir = filepath.Join(tempDir, filepath.Base(dir))
	return tempDir, nil
}

func (s *Supplier) InstallGems() error {
	if !s.appHasGemfile {
		return nil
	}

	s.warnBundleConfig()
	s.warnWindowsGemfile()

	tempDir, err := s.TempDir.CopyDirToTemp(s.Stager.BuildDir())
	if err != nil {
		return nil
	}
	gemfileLock, err := filepath.Rel(s.Stager.BuildDir(), s.Versions.Gemfile())
	if err != nil {
		return nil
	}
	gemfileLock = fmt.Sprintf("%s.lock", filepath.Join(tempDir, gemfileLock))

	if hasFile, err := s.Versions.HasWindowsGemfileLock(); err != nil {
		return err
	} else if hasFile {
		s.Log.Debug("Remove %s", gemfileLock)
		s.Log.Warning("Removing `Gemfile.lock` because it was generated on Windows.\nBundler will do a full resolve so native gems are handled properly.\nThis may result in unexpected gem versions being used in your app.\nIf you are using multi buildpacks, subsequent buildpacks may fail.\nIn rare occasions Bundler may not be able to resolve your dependencies at all.\nhttps://docs.cloudfoundry.org/buildpacks/ruby/windows.html")
		if err := os.Remove(gemfileLock); err != nil {
			return fmt.Errorf("Remove Gemfile.lock: %v", err)
		}
	}

	// Remove .bundle/config && copy if exists
	if exists, err := libbuildpack.FileExists(filepath.Join(tempDir, ".bundle", "config")); err != nil {
		return err
	} else if exists {
		os.Remove(filepath.Join(tempDir, ".bundle", "config"))
		libbuildpack.CopyFile(filepath.Join(s.Stager.BuildDir(), ".bundle", "config"), filepath.Join(tempDir, ".bundle", "config"))
	}

	vendorBundlePath, err := s.VendorBundlePath()
	if err != nil {
		return err
	}

	cmd := exec.Command("bundle", "config", "set", "path", filepath.Join(s.Stager.DepDir(), vendorBundlePath))
	cmd.Dir = tempDir
	if err := s.Command.Run(cmd); err != nil {
		return err
	}

	cmd = exec.Command("bundle", "config", "set", "without", os.Getenv("BUNDLE_WITHOUT"))
	cmd.Dir = tempDir
	if err := s.Command.Run(cmd); err != nil {
		return err
	}

	cmd = exec.Command("bundle", "config", "set", "bin", filepath.Join(s.Stager.DepDir(), "binstubs"))
	cmd.Dir = tempDir
	if err := s.Command.Run(cmd); err != nil {
		return err
	}

	args := []string{
		"install",
		"--jobs=4",
		"--retry=4",
	}

	if exists, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), "vendor", "cache")); err != nil {
		return err
	} else if exists {
		args = append(args, "--local")
	}

	if exists, err := libbuildpack.FileExists(gemfileLock); err != nil {
		return err
	} else if exists {
		cmd = exec.Command("bundle", "config", "set", "deployment", "true")
		cmd.Dir = tempDir
		if err := s.Command.Run(cmd); err != nil {
			return err
		}
	}

	bundlerVersion, err := s.Versions.GetBundlerVersion()
	if err != nil {
		return err
	}

	// Read Gemfile.lock to see if Bundler With version is >2, and not equal to the current bundler version
	// See: https://stackoverflow.com/questions/56680065/heroku-installs-bundler-then-throws-error-bundler-2-0-1
	if s.appHasGemfileLock {
		bundledWithVersion, err := s.Versions.BundledWithVersion()
		if err != nil {
			return fmt.Errorf("could not read Bundled With version from gemfile.lock: %s", err)
		}

		if bundledWithVersion != bundlerVersion && strings.HasPrefix(bundledWithVersion, "2") {
			if err := s.removeIncompatibleBundledWithVersion(bundledWithVersion); err != nil {
				return fmt.Errorf("could not remove Bundled With from end of "+
					"gemfile.lock: %s", err)
			}
		}
	}

	s.Log.BeginStep("Installing dependencies using bundler %s", bundlerVersion)
	s.Log.Info("Running: bundle %s", strings.Join(args, " "))

	env := os.Environ()
	env = append(env, "NOKOGIRI_USE_SYSTEM_LIBRARIES=true")

	cmd = exec.Command("bundle", args...)
	cmd.Dir = tempDir
	cmd.Stdout = text.NewIndentWriter(os.Stdout, []byte("       "))
	cmd.Stderr = text.NewIndentWriter(os.Stderr, []byte("       "))
	cmd.Env = env
	if err := s.Command.Run(cmd); err != nil {
		return err
	}

	cmd = exec.Command("bundle", "binstubs", "--all", "--force")
	cmd.Dir = tempDir
	if err := s.Command.Run(cmd); err != nil {
		return err
	}

	if err := s.regenerateBundlerBinStub(tempDir); err != nil {
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
			target := filepath.Join(s.Stager.DepDir(), "bin", file.Name())
			if exists, err := libbuildpack.FileExists(target); err != nil {
				return fmt.Errorf("Checking existence: %v", err)
			} else if !exists {
				source := filepath.Join(s.Stager.DepDir(), "binstubs", file.Name())
				if err := libbuildpack.CopyFile(source, target); err != nil {
					return fmt.Errorf("CopyFile: %v", err)
				}
			}
		}
	}

	// Save .bundle/config to global config
	if exists, err := libbuildpack.FileExists(filepath.Join(tempDir, ".bundle", "config")); err == nil && exists {
		s.Log.Debug("SaveBundleConfig; %s -> %s", filepath.Join(tempDir, ".bundle", "config"), os.Getenv("BUNDLE_CONFIG"))
		if err := os.Rename(filepath.Join(tempDir, ".bundle", "config"), os.Getenv("BUNDLE_CONFIG")); err != nil {
			return err
		}
	}

	// Save Gemfile.lock for finalize
	gemfileLockTarget := filepath.Join(s.Stager.DepDir(), "Gemfile.lock")
	if exists, err := libbuildpack.FileExists(gemfileLock); err == nil && exists {
		s.Log.Debug("SaveGemfileLock; %s -> %s", gemfileLock, gemfileLockTarget)
		if err := libbuildpack.CopyFile(gemfileLock, gemfileLockTarget); err != nil {
			return err
		}
	} else if err != nil {
		fmt.Printf("Error checking if Gemfile.lock exists: %v", err)
	}

	return os.RemoveAll(tempDir)
}

func (s *Supplier) removeIncompatibleBundledWithVersion(bundledWithVersion string) error {
	bundlerVersion, err := s.Versions.GetBundlerVersion()
	if err != nil {
		return err
	}

	s.Log.Warning(fmt.Sprintf(`Your Gemfile.lock was bundled with bundler %s, which is incompatible with the current bundler version (%s).`, bundledWithVersion, bundlerVersion))
	s.Log.Warning(`Deleting "Bundled With" from the Gemfile.lock`)

	gemfileLockPath := s.Versions.Gemfile() + ".lock"
	file, err := ioutil.ReadFile(gemfileLockPath)
	if err != nil {
		return fmt.Errorf("failed to read gemfile.lock: %s", err)
	}

	match := regexp.MustCompile(`BUNDLED WITH\s+(\w|\.|-)+\n`)
	output := match.ReplaceAll(file, []byte(""))

	return ioutil.WriteFile(gemfileLockPath, output, 0666)
}

func (s *Supplier) regenerateBundlerBinStub(appDir string) error {
	s.Log.BeginStep("Regenerating bundler binstubs...")
	cmd := exec.Command("bundle", "binstubs", "bundler", "--force", "--path", filepath.Join(s.Stager.DepDir(), "binstubs"))
	cmd.Dir = appDir
	cmd.Stdout = text.NewIndentWriter(os.Stdout, []byte("       "))
	cmd.Stderr = text.NewIndentWriter(os.Stderr, []byte("       "))
	if err := s.Command.Run(cmd); err != nil {
		return err
	}
	return libbuildpack.CopyFile(filepath.Join(s.Stager.DepDir(), "binstubs", "bundle"), filepath.Join(s.Stager.DepDir(), "bin", "bundle"))
}

func (s *Supplier) EnableLDLibraryPathEnv() error {
	if exists, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), "ld_library_path")); err != nil {
		return err
	} else if !exists {
		return nil
	}

	envVar := filepath.Join(s.Stager.BuildDir(), "ld_library_path")
	if env := os.Getenv("LD_LIBRARY_PATH"); env != "" {
		envVar += ":" + env
	}

	if err := os.Setenv("LD_LIBRARY_PATH", envVar); err != nil {
		return err
	}

	if err := s.Stager.WriteEnvFile("LD_LIBRARY_PATH", envVar); err != nil {
		return err
	}

	scriptContents := fmt.Sprintf(`export %[1]s="%[2]s$([[ ! -z "${%[1]s:-}" ]] && echo ":$%[1]s")"`, "LD_LIBRARY_PATH", filepath.Join("$HOME", "ld_library_path"))
	return s.Stager.WriteProfileD("app_lib_path.sh", scriptContents)
}

func (s *Supplier) CreateDefaultEnv() error {
	environmentDefaults := map[string]string{
		"RAILS_ENV":      "production",
		"RACK_ENV":       "production",
		"RAILS_GROUPS":   "assets",
		"BUNDLE_WITHOUT": "development:test",
		"BUNDLE_GEMFILE": "Gemfile",
		"BUNDLE_BIN":     filepath.Join(s.Stager.DepDir(), "binstubs"),
		"BUNDLE_CONFIG":  filepath.Join(s.Stager.DepDir(), "bundle_config"),
		"GEM_HOME":       filepath.Join(s.Stager.DepDir(), "gem_home"),
		"GEM_PATH": strings.Join([]string{
			filepath.Join(s.Stager.DepDir(), "bundler"),
			filepath.Join(s.Stager.DepDir(), "gem_home"),
		}, ":"),
	}

	return s.writeEnvFiles(environmentDefaults, false)
}

func (s *Supplier) AddPostRubyInstallDefaultEnv(engine string) error {
	rubyEngineVersion, err := s.Versions.RubyEngineVersion()
	if err != nil {
		return err
	}

	environmentDefaults := map[string]string{
		"GEM_PATH": strings.Join([]string{
			filepath.Join(s.Stager.DepDir(), "bundler"),
			filepath.Join(s.Stager.DepDir(), "vendor_bundle", engine, rubyEngineVersion),
			filepath.Join(s.Stager.DepDir(), "gem_home"),
		}, ":"),
	}
	s.Log.Debug("Setting post ruby install env: %v", environmentDefaults)
	return s.writeEnvFiles(environmentDefaults, true)
}

func (s *Supplier) AddPostRubyGemsInstallDefaultEnv(engine string) error {
	vendorBundlePath, err := s.VendorBundlePath()
	if err != nil {
		return err
	}

	environmentDefaults := map[string]string{
		"BUNDLE_PATH": filepath.Join(s.Stager.DepDir(), vendorBundlePath),
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

	vendorBundlePath := "vendor_bundle"

	depsIdx := s.Stager.DepsIdx()
	scriptContents := fmt.Sprintf(`
export LANG=${LANG:-en_US.UTF-8}
export RAILS_ENV=${RAILS_ENV:-production}
export RACK_ENV=${RACK_ENV:-production}
export RAILS_SERVE_STATIC_FILES=${RAILS_SERVE_STATIC_FILES:-enabled}
export RAILS_LOG_TO_STDOUT=${RAILS_LOG_TO_STDOUT:-enabled}

export GEM_HOME=${GEM_HOME:-$DEPS_DIR/%s/gem_home}
export GEM_PATH=${GEM_PATH:-$DEPS_DIR/%s/vendor_bundle/%s/%s:$DEPS_DIR/%s/gem_home:$DEPS_DIR/%s/bundler}

export BUNDLE_GEMFILE=${BUNDLE_GEMFILE:-$HOME/Gemfile}
export BUNDLE_PATH=$DEPS_DIR/%s/%s
export BUNDLE_BIN=$DEPS_DIR/%s/binstubs

## Change to current DEPS_DIR
bundle config PATH "$DEPS_DIR/%s/%s" > /dev/null
bundle config WITHOUT "%s" > /dev/null
bundle config BIN "$DEPS_DIR/%s/binstubs" > /dev/null
`, depsIdx,
		depsIdx, engine, rubyEngineVersion, depsIdx, depsIdx,
		depsIdx, vendorBundlePath,
		depsIdx,
		depsIdx, vendorBundlePath,
		os.Getenv("BUNDLE_WITHOUT"),
		depsIdx)

	if s.appHasGemfile && s.appHasGemfileLock {
		hasRails41, err := s.Versions.HasGemVersion("rails", ">=4.1.0.beta1")
		if err != nil {
			return fmt.Errorf("Could not determine rails version: %v", err)
		}
		if hasRails41 {
			metadata := s.Cache.Metadata()
			if metadata.SecretKeyBase == "" {
				metadata.SecretKeyBase, err = s.Command.Output(s.Stager.BuildDir(), "bundle", "exec", "rake", "secret")
				if err != nil {
					return fmt.Errorf("Failed to run 'rake secret': %v", err)
				}
				metadata.SecretKeyBase = strings.TrimSpace(metadata.SecretKeyBase)
			}
			scriptContents += fmt.Sprintf("\nexport SECRET_KEY_BASE=${SECRET_KEY_BASE:-%s}\n", metadata.SecretKeyBase)
		}
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
	if body, err := ioutil.ReadFile(s.Versions.Gemfile()); err == nil {
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

func (s *Supplier) installBundler(constraint string) error {
	version, err := libbuildpack.FindMatchingVersion(constraint, s.Manifest.AllDependencyVersions("bundler"))
	if err != nil {
		return fmt.Errorf("failure to install Bundler matching constraint, %s: %s", constraint, err)
	}

	if err := s.Installer.InstallDependency(libbuildpack.Dependency{Name: "bundler", Version: version}, filepath.Join(s.Stager.DepDir(), "bundler")); err != nil {
		return err
	}

	if err := s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "bundler", "bin"), "bin"); err != nil {
		return err
	}

	return nil
}
