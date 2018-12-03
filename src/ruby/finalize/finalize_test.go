package finalize_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/cloudfoundry/ruby-buildpack/src/ruby/finalize"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=finalize.go --destination=mocks_finalize_test.go --package=finalize_test

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Finalize")
}

var _ = Describe("Finalize", func() {
	var (
		err          error
		buildDir     string
		depsDir      string
		depsIdx      string
		finalizer    *finalize.Finalizer
		logger       *libbuildpack.Logger
		buffer       *bytes.Buffer
		mockCtrl     *gomock.Controller
		mockVersions *MockVersions
		mockCommand  *MockCommand
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "ruby-buildpack.build.")
		Expect(err).To(BeNil())

		depsDir, err = ioutil.TempDir("", "ruby-buildpack.deps.")
		Expect(err).To(BeNil())

		depsIdx = "9"
		Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)).To(Succeed())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockVersions = NewMockVersions(mockCtrl)
		mockCommand = NewMockCommand(mockCtrl)

		args := []string{buildDir, "", depsDir, depsIdx}
		stager := libbuildpack.NewStager(args, logger, &libbuildpack.Manifest{})

		finalizer = &finalize.Finalizer{
			Stager:   stager,
			Versions: mockVersions,
			Command:  mockCommand,
			Log:      logger,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	Describe("AssertGemfileLockExists", func() {
		Context("Gemfile.lock exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile.lock"), []byte("body"), 0644)).To(Succeed())
				Expect(filepath.Join(buildDir, "Gemfile.lock")).To(BeAnExistingFile())
			})
			It("Succeeds", func() {
				Expect(finalizer.AssertGemfileLockExists("Gemfile")).To(Succeed())
			})
		})
		Context("Gemfile.lock is missing", func() {
			BeforeEach(func() {
				Expect(filepath.Join(buildDir, "Gemfile.lock")).ToNot(BeAnExistingFile())
			})
			It("Fails", func() {
				Expect(finalizer.AssertGemfileLockExists("Gemfile")).To(MatchError("Gemfile.lock required"))
			})
		})
	})

	Describe("RestoreBundleConfig", func() {
		JustBeforeEach(func() {
			Expect(finalizer.RestoreBundleConfig()).To(Succeed())
		})

		Context("DEPS/IDX/.bundle_config exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "bundle_config"), []byte("bundler is awesome"), 0644)).To(Succeed())
			})

			It("Copies the config file to build dir", func() {
				Expect(filepath.Join(buildDir, ".bundle", "config")).To(BeARegularFile())
				Expect(ioutil.ReadFile(filepath.Join(buildDir, ".bundle", "config"))).To(Equal([]byte("bundler is awesome")))
			})
		})

		Context("DEPS/IDX/.bundle_config does NOT exist", func() {
			It("does nothing", func() {
				Expect(filepath.Join(buildDir, ".bundle", "config")).ToNot(BeARegularFile())
			})
		})
	})

	Describe("Install plugins", func() {
		JustBeforeEach(func() {
			Expect(finalizer.InstallPlugins()).To(Succeed())
		})

		Context("rails 3", func() {
			BeforeEach(func() {
				finalizer.RailsVersion = 3
			})

			Context("has rails_12factor gem", func() {
				BeforeEach(func() {
					finalizer.Gem12Factor = true
				})
				It("installs no plugins", func() {
					Expect(filepath.Join(buildDir, "vendor", "plugins")).ToNot(BeADirectory())
				})
			})

			Context("does not have rails_12factor gem", func() {
				BeforeEach(func() {
					finalizer.Gem12Factor = false
				})

				Context("the app has the gem rails_stdout_logging", func() {
					BeforeEach(func() {
						finalizer.GemStdoutLogging = true
					})

					It("does not install the plugin rails_log_stdout", func() {
						Expect(filepath.Join(buildDir, "vendor", "plugins", "rails_log_stdout")).ToNot(BeADirectory())
					})
				})

				Context("the app has the gem rails_serve_static_assets", func() {
					BeforeEach(func() {
						finalizer.GemStaticAssets = true
					})

					It("does not install the plugin rails3_serve_static_assets", func() {
						Expect(filepath.Join(buildDir, "vendor", "plugins", "rails3_serve_static_assets")).ToNot(BeADirectory())
					})
				})

				Context("the app has neither above gem", func() {
					It("installs plugin rails3_serve_static_assets", func() {
						Expect(filepath.Join(buildDir, "vendor", "plugins", "rails3_serve_static_assets", "init.rb")).To(BeARegularFile())
					})

					It("installs plugin rails_log_stdout", func() {
						Expect(filepath.Join(buildDir, "vendor", "plugins", "rails_log_stdout", "init.rb")).To(BeARegularFile())
					})
				})
			})
		})

		Context("rails 4", func() {
			var helpMessage string
			BeforeEach(func() {
				helpMessage = "Include 'rails_12factor' gem to enable all platform features"
				finalizer.RailsVersion = 4
			})

			It("installs no plugins", func() {
				Expect(filepath.Join(buildDir, "vendor", "plugins")).ToNot(BeADirectory())
			})

			Context("has rails_12factor gem", func() {
				BeforeEach(func() { finalizer.Gem12Factor = true })
				It("do not suggest rails_12factor to user", func() {
					Expect(buffer.String()).ToNot(ContainSubstring(helpMessage))
				})
			})

			Context("has rails_serve_static_assets and rails_stdout_logging gems", func() {
				BeforeEach(func() {
					finalizer.GemStdoutLogging = true
					finalizer.GemStaticAssets = true
				})
				It("do not suggest rails_12factor to user", func() {
					Expect(buffer.String()).ToNot(ContainSubstring(helpMessage))
				})
			})

			Context("has rails_serve_static_assets gem, but NOT rails_stdout_logging gem", func() {
				BeforeEach(func() {
					finalizer.GemStaticAssets = true
				})
				It("suggest rails_12factor to user", func() {
					Expect(buffer.String()).To(ContainSubstring(helpMessage))
				})
			})

			Context("has rails_stdout_logging gem, but NOT rails_serve_static_assets gem", func() {
				BeforeEach(func() {
					finalizer.GemStdoutLogging = true
				})
				It("suggest rails_12factor to user", func() {
					Expect(buffer.String()).To(ContainSubstring(helpMessage))
				})
			})

			Context("has none of the above gems", func() {
				It("suggest rails_12factor to user", func() {
					Expect(buffer.String()).To(ContainSubstring(helpMessage))
				})
			})
		})
		Context("rails 5", func() {
			BeforeEach(func() {
				finalizer.RailsVersion = 5
			})
			It("do not suggest anything", func() {
				Expect(buffer.String()).To(Equal(""))
			})
			It("installs no plugins", func() {
				Expect(filepath.Join(buildDir, "vendor", "plugins")).ToNot(BeADirectory())
			})
		})
	})

	Describe("create database.yml", func() {
		Context("config directory exists and activerecord < 4.1.0.beta", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(buildDir, "config"), 0755)).To(Succeed())
				mockVersions.EXPECT().HasGemVersion("activerecord", ">=4.1.0.beta").Return(false, nil)
			})

			It("writes config/database.yml", func() {
				finalizer.WriteDatabaseYml()
				Expect(filepath.Join(buildDir, "config", "database.yml")).To(BeARegularFile())
			})

			It("logs topic", func() {
				finalizer.WriteDatabaseYml()
				Expect(buffer.String()).To(ContainSubstring("Writing config/database.yml to read from DATABASE_URL"))
			})
		})

		Context("config directory does not exists", func() {
			It("does not write config/database.yml", func() {
				finalizer.WriteDatabaseYml()
				Expect(filepath.Join(buildDir, "config", "database.yml")).ToNot(BeAnExistingFile())
			})

			It("does not log", func() {
				finalizer.WriteDatabaseYml()
				Expect(buffer.String()).To(BeEmpty())
			})
		})

		Context("config directory exists, but activerecord >= 4.1.0.beta", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(buildDir, "config"), 0755)).To(Succeed())
				mockVersions.EXPECT().HasGemVersion("activerecord", ">=4.1.0.beta").Return(true, nil)
			})

			It("does not write config/database.yml", func() {
				finalizer.WriteDatabaseYml()
				Expect(filepath.Join(buildDir, "config", "database.yml")).ToNot(BeAnExistingFile())
			})

			It("does not log", func() {
				finalizer.WriteDatabaseYml()
				Expect(buffer.String()).To(BeEmpty())
			})
		})
	})

	Describe("PrecompileAssets", func() {
		Context("app does not have assets:precompile task", func() {
			It("doesn't run assets:precompile", func() {
				mockCommand.EXPECT().Run(gomock.Any()).Do(func(cmd *exec.Cmd) {
					Expect(cmd.Args).To(Equal([]string{"bundle", "exec", "rake", "-n", "assets:precompile"}))
				}).Return(errors.New("app does not have assets:precompile task"))
				Expect(finalizer.PrecompileAssets()).To(Succeed())
			})
		})

		Context("app has assets:precompile task", func() {
			var cmds []*exec.Cmd
			BeforeEach(func() {
				mockVersions.EXPECT().HasGem(gomock.Any()).AnyTimes().Return(false, nil)
				cmds = []*exec.Cmd{}
				mockCommand.EXPECT().Run(gomock.Any()).AnyTimes().Do(func(cmd *exec.Cmd) {
					cmds = append(cmds, cmd)
				}).Return(nil)
			})
			Context("Rails < 4", func() {
				BeforeEach(func() {
					finalizer.RailsVersion = 3
				})
				Context("public/assets/manifest.yml is present", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(buildDir, "public", "assets"), 0755)).To(Succeed())
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "public", "assets", "manifest.yml"), []byte("memanifest"), 0644)).To(Succeed())
					})
					It("skips assets:precompile", func() {
						Expect(finalizer.PrecompileAssets()).To(Succeed())
						Expect(cmds).To(BeEmpty())
						Expect(buffer.String()).To(ContainSubstring("Detected assets manifest file, assuming assets were compiled locally"))
					})
				})
				Context("public/assets/manifest.yml is not present", func() {
					BeforeEach(func() {
						Expect(libbuildpack.FileExists(filepath.Join(buildDir, "public", "assets", "manifest.yml"))).To(BeFalse())
					})
					It("runs assets:precompile with DATABASE_URL", func() {
						Expect(finalizer.PrecompileAssets()).To(Succeed())
						Expect(cmds).To(HaveLen(3))
						Expect(cmds[1].Args).To(Equal([]string{"bundle", "exec", "rake", "assets:precompile"}))
						Expect(cmds[1].Env).To(ContainElement("DATABASE_URL=://user:pass@127.0.0.1/dbname"))
					})
				})
			})
			Context("Rails >= 4", func() {
				BeforeEach(func() {
					finalizer.RailsVersion = 4
				})
				Context("public/assets/.sprockets-manifest-*.json is present", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(buildDir, "public", "assets"), 0755)).To(Succeed())
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "public", "assets", ".sprockets-manifest-123.json"), []byte("memanifest"), 0644)).To(Succeed())
					})
					It("skips assets:precompile", func() {
						Expect(finalizer.PrecompileAssets()).To(Succeed())
						Expect(cmds).To(BeEmpty())
						Expect(buffer.String()).To(ContainSubstring("Detected assets manifest file, assuming assets were compiled locally"))
					})
				})
				Context("public/assets/manifest-*.json is present", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(buildDir, "public", "assets"), 0755)).To(Succeed())
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "public", "assets", "manifest-123.json"), []byte("memanifest"), 0644)).To(Succeed())
					})
					It("skips assets:precompile", func() {
						Expect(finalizer.PrecompileAssets()).To(Succeed())
						Expect(cmds).To(BeEmpty())
						Expect(buffer.String()).To(ContainSubstring("Detected assets manifest file, assuming assets were compiled locally"))
					})
				})
				Context("public/assets/manifest.yml is present", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(buildDir, "public", "assets"), 0755)).To(Succeed())
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "public", "assets", "manifest.yml"), []byte("memanifest"), 0644)).To(Succeed())
					})
					It("runs assets:precompile with DATABASE_URL", func() {
						Expect(finalizer.PrecompileAssets()).To(Succeed())
						Expect(cmds).To(HaveLen(4))
						Expect(cmds[1].Args).To(Equal([]string{"bundle", "exec", "rake", "assets:precompile"}))
						Expect(cmds[1].Env).To(ContainElement("DATABASE_URL=://user:pass@127.0.0.1/dbname"))

						Expect(cmds[3].Args).To(Equal([]string{"bundle", "exec", "rake", "assets:clean"}))
					})
				})
				Context("No manifest files exist in public/assets/", func() {
					It("runs assets:precompile with DATABASE_URL", func() {
						Expect(finalizer.PrecompileAssets()).To(Succeed())
						Expect(cmds).To(HaveLen(4))
						Expect(cmds[1].Args).To(Equal([]string{"bundle", "exec", "rake", "assets:precompile"}))
						Expect(cmds[1].Env).To(ContainElement("DATABASE_URL=://user:pass@127.0.0.1/dbname"))

						Expect(cmds[3].Args).To(Equal([]string{"bundle", "exec", "rake", "assets:clean"}))
					})
				})

				findAllWithPrefix := func(prefix string, inp []string) []string {
					var out []string
					for _, s := range inp {
						if strings.HasPrefix(s, prefix) {
							out = append(out, s)
						}
					}
					return out
				}

				Context("SECRET_KEY_BASE is set", func() {
					BeforeEach(func() { os.Setenv("SECRET_KEY_BASE", "existing-key") })
					AfterEach(func() { os.Unsetenv("SECRET_KEY_BASE") })
					It("passes SECRET_KEY_BASE through", func() {
						Expect(finalizer.PrecompileAssets()).To(Succeed())
						Expect(cmds).To(HaveLen(4))
						Expect(cmds[1].Args).To(Equal([]string{"bundle", "exec", "rake", "assets:precompile"}))
						Expect(findAllWithPrefix("SECRET_KEY_BASE=", cmds[1].Env)).To(Equal([]string{"SECRET_KEY_BASE=existing-key"}))
					})
				})
				Context("SECRET_KEY_BASE is NOT set", func() {
					It("sets a dummy key", func() {
						Expect(finalizer.PrecompileAssets()).To(Succeed())
						Expect(cmds).To(HaveLen(4))
						Expect(cmds[1].Args).To(Equal([]string{"bundle", "exec", "rake", "assets:precompile"}))
						Expect(findAllWithPrefix("SECRET_KEY_BASE=", cmds[1].Env)).To(Equal([]string{"SECRET_KEY_BASE=dummy-staging-key"}))
					})
				})
			})
		})
	})

	Describe("best practice warnings", func() {
		Context("RAILS_ENV == production", func() {
			BeforeEach(func() { os.Setenv("RAILS_ENV", "production") })
			AfterEach(func() { os.Setenv("RAILS_ENV", "") })

			It("does not warn the user", func() {
				finalizer.BestPracticeWarnings()
				Expect(buffer.String()).To(Equal(""))
			})
		})

		Context("RAILS_ENV != production", func() {
			BeforeEach(func() { os.Setenv("RAILS_ENV", "otherenv") })
			AfterEach(func() { os.Setenv("RAILS_ENV", "") })

			It("warns the user", func() {
				finalizer.BestPracticeWarnings()
				Expect(buffer.String()).To(ContainSubstring("You are deploying to a non-production environment: otherenv"))
			})
		})
	})

	Describe("DeleteVendorBundle", func() {
		Context("vendor/bundle in pushed app", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(buildDir, "vendor", "bundle"), 0755)).To(Succeed())
			})

			It("warns about the presence of the directory", func() {
				finalizer.DeleteVendorBundle()
				Expect(buffer.String()).To(ContainSubstring("**WARNING** Removing `vendor/bundle`."))
				Expect(buffer.String()).To(ContainSubstring("Checking in `vendor/bundle` is not supported. Please remove this directory and add it to your .gitignore. To vendor your gems with Bundler, use `bundle pack` instead."))
			})

			It("deletes the directory", func() {
				finalizer.DeleteVendorBundle()
				Expect(filepath.Join(buildDir, "vendor", "bundle")).ToNot(BeADirectory())
			})
		})

		Context("vendor/bundle not in pushed app", func() {
			It("does not warn about the presence of the directory", func() {
				finalizer.DeleteVendorBundle()
				Expect(buffer.String()).ToNot(ContainSubstring("**WARNING** Removing `vendor/bundle`."))
			})
		})
	})

	Describe("CopyToAppBin", func() {
		BeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(buildDir, "bin"), 0755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "binstubs"), 0755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "bin"), 0755)).To(Succeed())
		})
		Context("file exists in app/bin", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "bin", "rake"), []byte("original"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "binstubs", "rake"), []byte("dep/binstub"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "bin", "rake"), []byte("dep/bin"), 0755)).To(Succeed())
			})

			It("remains unchanged", func() {
				Expect(finalizer.CopyToAppBin()).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(buildDir, "bin", "rake"))).To(ContainSubstring("original"))
			})
		})

		Context("file exists in dep/bin and dep/binstubs", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "binstubs", "rake"), []byte("dep/binstub"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "bin", "rake"), []byte("dep/bin"), 0755)).To(Succeed())
			})

			It("copies deps binstubs file", func() {
				Expect(finalizer.CopyToAppBin()).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(buildDir, "bin", "rake"))).To(ContainSubstring("dep/binstub"))
			})
		})

		Context("file only exists in dep/bin", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "bin", "rake"), []byte("dep/bin"), 0755)).To(Succeed())
			})

			It("creates a shim for dep/bin", func() {
				Expect(finalizer.CopyToAppBin()).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(buildDir, "bin", "rake"))).To(ContainSubstring(`Kernel.exec "#{ENV['DEPS_DIR']}/%s/bin/rake", *ARGV`, depsIdx))
			})
		})

		Context("binstubs does not exist", func() {
			BeforeEach(func() {
				Expect(os.RemoveAll(filepath.Join(depsDir, depsIdx, "binstubs"))).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "bin", "rake"), []byte("dep/bin"), 0755)).To(Succeed())
			})
			It("creates a shim for dep/bin", func() {
				Expect(finalizer.CopyToAppBin()).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(buildDir, "bin", "rake"))).To(ContainSubstring(`Kernel.exec "#{ENV['DEPS_DIR']}/%s/bin/rake", *ARGV`, depsIdx))
			})
		})
	})
})
