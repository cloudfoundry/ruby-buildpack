package finalize

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/kr/text"
)

type Stager interface {
	BuildDir() string
	DepsIdx() string
	DepDir() string
}

type Command interface {
	Run(*exec.Cmd) error
}

type Finalizer struct {
	Stager           Stager
	Versions         Versions
	Log              *libbuildpack.Logger
	Command          Command
	Gem12Factor      bool
	GemStaticAssets  bool
	GemStdoutLogging bool
	RailsVersion     int
}

func Run(f *Finalizer) error {
	f.Log.BeginStep("Finalizing Ruby")

	if err := f.Setup(); err != nil {
		f.Log.Error("Error determining versions: %v", err)
		return err
	}

	if err := f.RestoreGemfileLock(); err != nil {
		f.Log.Error("Error copying Gemfile.lock to app: %v", err)
		return err
	}

	if err := f.RestoreBundleConfig(); err != nil {
		f.Log.Error("Error copying .bundle/config to app: %v", err)
		return err
	}

	if err := f.InstallPlugins(); err != nil {
		f.Log.Error("Error installing plugins: %v", err)
		return err
	}

	if err := f.WriteDatabaseYml(); err != nil {
		f.Log.Error("Error writing database.yml: %v", err)
		return err
	}

	if err := f.PrecompileAssets(); err != nil {
		f.Log.Error("Error precompiling assets: %v", err)
		return err
	}

	f.BestPracticeWarnings()

	if err := f.DeleteVendorBundle(); err != nil {
		f.Log.Error("Error deleting vendor/bundle: %v", err)
		return err
	}

	if err := f.CopyToAppBin(); err != nil {
		f.Log.Error("Error creating files in bin: %v", err)
		return err
	}

	data, err := f.GenerateReleaseYaml()
	if err != nil {
		f.Log.Error("Error generating release YAML: %v", err)
		return err
	}
	releasePath := filepath.Join(f.Stager.BuildDir(), "tmp", "ruby-buildpack-release-step.yml")
	libbuildpack.NewYAML().Write(releasePath, data)

	return nil
}

func (f *Finalizer) Setup() error {
	var err error

	f.Gem12Factor, err = f.Versions.HasGem("rails_12factor")
	if err != nil {
		return err
	}

	f.GemStdoutLogging, err = f.Versions.HasGem("rails_stdout_logging")
	if err != nil {
		return err
	}

	f.GemStaticAssets, err = f.Versions.HasGem("rails_serve_static_assets")
	if err != nil {
		return err
	}

	f.RailsVersion, err = f.Versions.GemMajorVersion("rails")
	if err != nil {
		return err
	}

	return nil
}

func (f *Finalizer) RestoreGemfileLock() error {
	source := filepath.Join(f.Stager.DepDir(), "Gemfile.lock")
	f.Log.Debug("RestoreGemfileLock; %s", source)
	if exists, err := libbuildpack.FileExists(source); err != nil {
		return err
	} else if exists {
		gemfile := "Gemfile"
		if os.Getenv("BUNDLE_GEMFILE") != "" {
			gemfile = os.Getenv("BUNDLE_GEMFILE")
		}
		target := filepath.Join(f.Stager.BuildDir(), gemfile) + ".lock"
		f.Log.Debug("RestoreGemfileLock; exists, copy to %s", target)
		return os.Rename(source, target)
	}
	return nil
}

func (f *Finalizer) RestoreBundleConfig() error {
	source := filepath.Join(f.Stager.DepDir(), "bundle_config")
	f.Log.Debug("RestoreBundleConfig; %s", source)
	if exists, err := libbuildpack.FileExists(source); err != nil {
		return err
	} else if exists {
		target := filepath.Join(f.Stager.BuildDir(), ".bundle", "config")
		f.Log.Debug("RestoreBundleConfig; exists, copy to %s", target)
		os.MkdirAll(filepath.Join(f.Stager.BuildDir(), ".bundle"), 0755)
		return os.Rename(source, target)
	}
	return nil
}

func (f *Finalizer) WriteDatabaseYml() error {
	if exists, err := libbuildpack.FileExists(filepath.Join(f.Stager.BuildDir(), "config")); err != nil {
		return err
	} else if !exists {
		return nil
	}
	if rails41Plus, err := f.Versions.HasGemVersion("activerecord", ">=4.1.0.beta"); err != nil {
		return err
	} else if rails41Plus {
		return nil
	}

	f.Log.BeginStep("Writing config/database.yml to read from DATABASE_URL")
	if err := ioutil.WriteFile(filepath.Join(f.Stager.BuildDir(), "config", "database.yml"), []byte(config_database_yml), 0644); err != nil {
		return err
	}

	return nil
}

func (f *Finalizer) databaseUrl() string {
	if os.Getenv("DATABASE_URL") != "" {
		return os.Getenv("DATABASE_URL")
	}

	gems := map[string]string{
		"pg":            "postgres",
		"jdbc-postgres": "postgres",
		"mysql":         "mysql",
		"mysql2":        "mysql2",
		"sqlite3":       "sqlite3",
		"sqlite3-ruby":  "sqlite3",
	}

	scheme := ""
	for k, v := range gems {
		if a, err := f.Versions.HasGem(k); err == nil && a {
			scheme = v
			break
		}
	}

	return fmt.Sprintf("%s://user:pass@127.0.0.1/dbname", scheme)
}

func (f *Finalizer) hasPrecompiledAssets() (bool, error) {
	globs := []string{".sprockets-manifest-*.json", "manifest-*.json"}
	if f.RailsVersion < 4 {
		globs = []string{"manifest.yml"}
	}
	for _, glob := range globs {
		if matches, err := filepath.Glob(filepath.Join(f.Stager.BuildDir(), "public", "assets", glob)); err != nil {
			return false, err
		} else if len(matches) > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (f *Finalizer) PrecompileAssets() error {
	if exists, err := f.hasPrecompiledAssets(); err != nil {
		return err
	} else if exists {
		f.Log.Info("Detected assets manifest file, assuming assets were compiled locally")
		return nil
	}

	cmd := exec.Command("bundle", "exec", "rake", "-n", "assets:precompile")
	cmd.Dir = f.Stager.BuildDir()
	if err := f.Command.Run(cmd); err != nil {
		return nil
	}

	env := append(os.Environ(), fmt.Sprintf("DATABASE_URL=%s", f.databaseUrl()))

	f.Log.BeginStep("Precompiling assets")
	startTime := time.Now()
	cmd = exec.Command("bundle", "exec", "rake", "assets:precompile")
	cmd.Dir = f.Stager.BuildDir()
	cmd.Stdout = text.NewIndentWriter(os.Stdout, []byte("       "))
	cmd.Stderr = text.NewIndentWriter(os.Stderr, []byte("       "))
	cmd.Env = env
	err := f.Command.Run(cmd)

	f.Log.Info("Asset precompilation completed (%v)", time.Since(startTime))

	if f.RailsVersion >= 4 && err == nil {
		f.Log.Info("Cleaning assets")
		cmd = exec.Command("bundle", "exec", "rake", "assets:clean")
		cmd.Dir = f.Stager.BuildDir()
		cmd.Stdout = text.NewIndentWriter(os.Stdout, []byte("       "))
		cmd.Stderr = text.NewIndentWriter(os.Stderr, []byte("       "))
		cmd.Env = env
		err = f.Command.Run(cmd)
	}

	return err
}

func (f *Finalizer) InstallPlugins() error {
	if f.Gem12Factor {
		return nil
	}

	if f.RailsVersion == 4 {
		if !(f.GemStdoutLogging && f.GemStaticAssets) {
			f.Log.Protip("Include 'rails_12factor' gem to enable all platform features", "https://devcenter.heroku.com/articles/rails-integration-gems")
		}
		return nil
	}

	if f.RailsVersion == 2 || f.RailsVersion == 3 {
		if err := f.installPluginStdoutLogger(); err != nil {
			return err
		}
		if err := f.installPluginServeStaticAssets(); err != nil {
			return err
		}
	}
	return nil
}

func (f *Finalizer) installPluginStdoutLogger() error {
	if f.GemStdoutLogging {
		return nil
	}

	f.Log.BeginStep("Injecting plugin 'rails_log_stdout'")

	code := `
begin
  STDOUT.sync = true
  def Rails.cloudfoundry_stdout_logger
    logger = Logger.new(STDOUT)
    logger = ActiveSupport::TaggedLogging.new(logger) if defined?(ActiveSupport::TaggedLogging)
    level = ENV['LOG_LEVEL'].to_s.upcase
    level = 'INFO' unless %w[DEBUG INFO WARN ERROR FATAL UNKNOWN].include?(level)
    logger.level = Logger.const_get(level)
    logger
  end
  Rails.logger = Rails.application.config.logger = Rails.cloudfoundry_stdout_logger
rescue Exception => ex
  puts %Q{WARNING: Exception during rails_log_stdout init: #{ex.message}}
end
`

	if err := os.MkdirAll(filepath.Join(f.Stager.BuildDir(), "vendor", "plugins", "rails_log_stdout"), 0755); err != nil {
		return fmt.Errorf("Error creating rails_log_stdout plugin directory: %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(f.Stager.BuildDir(), "vendor", "plugins", "rails_log_stdout", "init.rb"), []byte(code), 0644); err != nil {
		return fmt.Errorf("Error writing rails_log_stdout plugin file: %v", err)
	}
	return nil
}

func (f *Finalizer) installPluginServeStaticAssets() error {
	if f.GemStaticAssets {
		return nil
	}

	f.Log.BeginStep("Injecting plugin 'rails3_serve_static_assets'")

	code := "Rails.application.class.config.serve_static_assets = true\n"

	if err := os.MkdirAll(filepath.Join(f.Stager.BuildDir(), "vendor", "plugins", "rails3_serve_static_assets"), 0755); err != nil {
		return fmt.Errorf("Error creating rails3_serve_static_assets plugin directory: %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(f.Stager.BuildDir(), "vendor", "plugins", "rails3_serve_static_assets", "init.rb"), []byte(code), 0644); err != nil {
		return fmt.Errorf("Error writing rails3_serve_static_assets plugin file: %v", err)
	}
	return nil
}

func (f *Finalizer) BestPracticeWarnings() {
	if os.Getenv("RAILS_ENV") != "production" {
		f.Log.Warning("You are deploying to a non-production environment: %s", os.Getenv("RAILS_ENV"))
	}
}

func (f *Finalizer) DeleteVendorBundle() error {
	if exists, err := libbuildpack.FileExists(filepath.Join(f.Stager.BuildDir(), "vendor", "bundle")); err != nil {
		return err
	} else if exists {
		f.Log.Warning("Removing `vendor/bundle`.\nChecking in `vendor/bundle` is not supported. Please remove this directory and add it to your .gitignore. To vendor your gems with Bundler, use `bundle pack` instead.")
		return os.RemoveAll(filepath.Join(f.Stager.BuildDir(), "vendor", "bundle"))
	}

	return nil
}

func (f *Finalizer) CopyToAppBin() error {
	f.Log.BeginStep("Copy binaries to app/bin directory")

	binDir := filepath.Join(f.Stager.BuildDir(), "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("Could not create /app/bin directory: %v", err)
	}

	files, err := ioutil.ReadDir(filepath.Join(f.Stager.DepDir(), "binstubs"))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Could not read dep/binstubs directory: %v", err)
		}
		files = []os.FileInfo{}
	}
	for _, file := range files {
		source := filepath.Join(f.Stager.DepDir(), "binstubs", file.Name())
		target := filepath.Join(binDir, file.Name())
		if exists, err := libbuildpack.FileExists(target); err != nil {
			return fmt.Errorf("Checking existence: %v", err)
		} else if !exists {
			if err := libbuildpack.CopyFile(source, target); err != nil {
				return fmt.Errorf("CopyFile: %v", err)
			}
		}
	}

	files, err = ioutil.ReadDir(filepath.Join(f.Stager.DepDir(), "bin"))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Could not read dep/bin directory: %v", err)
		}
		files = []os.FileInfo{}
	}
	for _, file := range files {
		target := filepath.Join(binDir, file.Name())
		if exists, err := libbuildpack.FileExists(target); err != nil {
			return fmt.Errorf("Checking existence: %v", err)
		} else if !exists {
			contents := fmt.Sprintf("#!/bin/bash\nexec $DEPS_DIR/%s/bin/%s \"$@\"\n", f.Stager.DepsIdx(), file.Name())
			if err := ioutil.WriteFile(target, []byte(contents), 0755); err != nil {
				return fmt.Errorf("WriteFile: %v", err)
			}
		}
	}
	return nil
}
