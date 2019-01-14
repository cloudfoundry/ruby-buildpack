package versions

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Manifest interface {
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Versions struct {
	buildDir       string
	depDir         string
	manifest       Manifest
	cachedSpecs    map[string]string
	bundlerVersion string
}

func New(buildDir string, depDir string, manifest Manifest) *Versions {
	bundlerVersions := manifest.AllDependencyVersions("bundler")
	bundlerVersion := ""
	if len(bundlerVersions) > 0 {
		bundlerVersion = bundlerVersions[len(bundlerVersions)-1]
	}
	return &Versions{
		buildDir:       buildDir,
		depDir:         depDir,
		manifest:       manifest,
		bundlerVersion: bundlerVersion,
	}
}

type output struct {
	Error  string      `json:"error"`
	Output interface{} `json:"output"`
}

func (v *Versions) SetBundlerVersion(version string) {
	v.bundlerVersion = version
}

func (v *Versions) GetBundlerVersion() string {
	return v.bundlerVersion
}

func (v *Versions) CheckRubyAndBundlerVersions() (bool, error) {
	engine, err := v.Engine()
	if err != nil {
		return false, err
	}

	if engine == "jruby" {
		return true, nil
	}

	gemfileRubyVersion, err := v.Version()
	if err != nil {
		return false, err
	}

	if gemfileRubyVersion == "" {
		dep, err := v.manifest.DefaultVersion("ruby")
		if err != nil {
			return false, err
		}
		gemfileRubyVersion = dep.Version
	}

	rubyConstraint, err := semver.NewConstraint("<= 2.2.X")
	if err != nil {
		return false, err
	}

	bundlerConstraint, err := semver.NewConstraint(">= 2.X.X")
	if err != nil {
		return false, err
	}

	bundlerSemver, err := semver.NewVersion(v.bundlerVersion)
	if err != nil {
		return false, err
	}

	rubySemver, err := semver.NewVersion(gemfileRubyVersion)
	if err != nil {
		return false, err
	}

	if engine == "ruby" && rubyConstraint.Check(rubySemver) && bundlerConstraint.Check(bundlerSemver) {
		return false, nil
	}

	return true, nil
}

func (v *Versions) Engine() (string, error) {
	gemfile := v.Gemfile()
	code := fmt.Sprintf(`
		b = Bundler::Dsl.evaluate('%s', '%s.lock', {}).ruby_version if File.exists?('%s')
	  return 'ruby' if !b
		b.engine
	`, filepath.Base(gemfile), filepath.Base(gemfile), filepath.Base(gemfile))

	data, err := v.run(filepath.Dir(gemfile), code, []string{})
	if err != nil {
		return "", err
	}

	return data.(string), nil
}

func (v *Versions) Version() (string, error) {
	versions := v.manifest.AllDependencyVersions("ruby")
	gemfile := v.Gemfile()
	code := fmt.Sprintf(`
		b = Bundler::Dsl.evaluate('%s', '%s.lock', {}).ruby_version
	  return '' if !b

		r = Gem::Requirement.create(b.versions)
		version = input.select { |v| r.satisfied_by? Gem::Version.new(v) }.sort.last
		raise "No Matching versions, ruby #{r} not found in this buildpack" unless version
		version
	`, filepath.Base(gemfile), filepath.Base(gemfile))

	data, err := v.run(filepath.Dir(gemfile), code, versions)
	if err != nil {
		return "", err
	}

	return data.(string), nil
}

func (v *Versions) JrubyVersion() (string, error) {
	gemfile := v.Gemfile()
	code := fmt.Sprintf(`
		b = Bundler::Dsl.evaluate('%s', '%s.lock', {}).ruby_version
	  return '' if !b

	  "#{b.versions_string(b.engine_versions)}"
	`, filepath.Base(gemfile), filepath.Base(gemfile))

	data, err := v.run(filepath.Dir(gemfile), code, []string{})
	if err != nil {
		return "", err
	}

	return data.(string), nil
}

func (v *Versions) RubyEngineVersion() (string, error) {
	code := `require 'rbconfig';RbConfig::CONFIG['ruby_version']`

	data, err := v.run(v.buildDir, code, []string{})
	if err != nil {
		return "", err
	}
	return data.(string), nil
}

func (v *Versions) VersionConstraint(version string, constraints ...string) (bool, error) {
	code := `
		version = input.shift
		Gem::Requirement.create(input).satisfied_by? Gem::Version.new(version)
	`

	data, err := v.run(v.buildDir, code, append([]string{version}, constraints...))
	if err != nil {
		return false, err
	}

	return data.(bool), nil
}

func (v *Versions) HasGemVersion(gem string, constraints ...string) (bool, error) {
	specs, err := v.specs()
	if err != nil {
		return false, err
	}
	if specs[gem] == "" {
		return false, nil
	}

	return v.VersionConstraint(specs[gem], constraints...)
}

func (v *Versions) HasGem(gem string) (bool, error) {
	specs, err := v.specs()
	if err != nil {
		return false, err
	}
	if specs[gem] != "" {
		return true, nil
	}
	return false, nil
}

func (v *Versions) GemMajorVersion(gem string) (int, error) {
	specs, err := v.specs()
	if err != nil {
		return -1, err
	}
	if specs[gem] == "" {
		return -1, nil
	}

	code := `Gem::Version.new(input.first).segments.first.to_s`
	data, err := v.run(v.buildDir, code, []string{specs[gem]})
	if err != nil {
		return -1, err
	}

	if i, err := strconv.Atoi(data.(string)); err == nil {
		return i, nil
	} else {
		return -1, err
	}
}

//Should return true if either:
// (1) the only platform in the Gemfile.lock is windows (mingw/mswin)
//     (there is no ruby or jruby platform)
//     -or-
// (2) the Gemfile.lock line endings are /r/n, rather than just /n
func (v *Versions) HasWindowsGemfileLock() (bool, error) {
	gemfileLockPath := v.Gemfile() + ".lock"
	if good, err := libbuildpack.FileExists(gemfileLockPath); err != nil {
		return false, err
	} else if !good {
		return false, nil
	}
	if bytes, err := ioutil.ReadFile(gemfileLockPath); err != nil {
		return false, err
	} else if strings.Contains(string(bytes), "\r\n") {
		return true, nil
	}

	code := `
	  return false if !File.exists?(input["gemfilelock"])
		parsed = Bundler::LockfileParser.new(File.read(input["gemfilelock"]))
		parsed.platforms.detect do |platform|
		  if platform.is_a?(String)
        /ruby/.match(platform)
			elsif platform.is_a?(Gem::Platform)
			  /java/.match(platform.os)
			end
    end.nil?
	`

	data, err := v.run(filepath.Dir(v.Gemfile()),
		code,
		map[string]string{"gemfilelock": fmt.Sprintf("%s.lock", v.Gemfile())})
	if err != nil {
		return false, err
	}
	return data.(bool), nil
}

func (v *Versions) specs() (map[string]string, error) {
	if len(v.cachedSpecs) > 0 {
		return v.cachedSpecs, nil
	}
	code := `
		parsed = Bundler::LockfileParser.new(File.read(input["gemfilelock"]))
		Hash[*(parsed.specs.map{|spec| [spec.name, spec.version.to_s]}).flatten]
	`

	data, err := v.run(filepath.Dir(v.Gemfile()), code, map[string]string{"gemfilelock": fmt.Sprintf("%s.lock", v.Gemfile())})
	if err != nil {
		return nil, err
	}
	specs := make(map[string]string, 0)
	for k, v := range data.(map[string]interface{}) {
		specs[k] = v.(string)
	}
	v.cachedSpecs = specs
	return v.cachedSpecs, nil
}

func (v *Versions) Gemfile() string {
	gemfile := "Gemfile"
	if os.Getenv("BUNDLE_GEMFILE") != "" {
		gemfile = os.Getenv("BUNDLE_GEMFILE")
	}
	return filepath.Join(v.buildDir, gemfile)
}

func (v *Versions) run(dir, code string, in interface{}) (interface{}, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return "", err
	}

	code = fmt.Sprintf(`
	  stdout, $stdout = $stdout, $stderr
		begin
			def data(input)
				%s
			end
			input = JSON.parse(STDIN.read)
			out = data(input)
			stdout.puts({error:nil, data:out}.to_json)
		rescue => e
			stdout.puts({error:e.to_s, data:nil}.to_json)
		end
	`, code)

	args := []string{"-rjson", "-e", code}
	if strings.Contains(code, "Bundler::") {
		bundlerPath := filepath.Join(v.depDir, "bundler", "gems", fmt.Sprintf("bundler-%s", v.bundlerVersion), "lib")
		args = append([]string{fmt.Sprintf("-I%s", bundlerPath), "-rbundler"}, args...)
	}

	cmd := exec.Command("ruby", args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(string(data))
	body, err := cmd.Output()
	if err != nil {
		return "", err
	}

	output := struct {
		Error string      `json:"error"`
		Data  interface{} `json:"data"`
	}{}
	if err := json.Unmarshal(body, &output); err != nil {
		return "", err
	}
	if output.Error != "" {
		return "", fmt.Errorf("Running ruby: %s", output.Error)
	}
	return output.Data, nil
}
