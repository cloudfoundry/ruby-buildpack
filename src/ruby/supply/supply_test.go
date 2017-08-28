package supply_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"ruby/cache"
	"ruby/supply"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	gomock "github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=supply.go --destination=mocks_test.go --package=supply_test

var _ = Describe("Supply", func() {
	var (
		err          error
		buildDir     string
		depsDir      string
		depsIdx      string
		supplier     *supply.Supplier
		logger       *libbuildpack.Logger
		buffer       *bytes.Buffer
		mockCtrl     *gomock.Controller
		mockManifest *MockManifest
		mockVersions *MockVersions
		mockCommand  *MockCommand
		mockCache    *MockCache
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
		mockManifest = NewMockManifest(mockCtrl)
		mockVersions = NewMockVersions(mockCtrl)
		mockCommand = NewMockCommand(mockCtrl)
		mockCache = NewMockCache(mockCtrl)

		args := []string{buildDir, "", depsDir, depsIdx}
		stager := libbuildpack.NewStager(args, logger, &libbuildpack.Manifest{})

		supplier = &supply.Supplier{
			Stager:   stager,
			Manifest: mockManifest,
			Log:      logger,
			Versions: mockVersions,
			Cache:    mockCache,
			Command:  mockCommand,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	PIt("InstallBundler", func() {})
	PIt("InstallNode", func() {})
	PIt("InstallRuby", func() {})

	Describe("CalcChecksum", func() {
		BeforeEach(func() {
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte("source \"https://rubygems.org\"\r\ngem \"rack\"\r\n"), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "other"), []byte("other"), 0644)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(buildDir, "dir"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "dir", "other"), []byte("other"), 0644)).To(Succeed())
		})
		It("Returns an MD5 of the full contents", func() {
			Expect(supplier.CalcChecksum()).To(Equal("d8be25466f8d12112d354e1a4add36a3"))
		})

		Context(".cloudfoundry directory", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(buildDir, ".cloudfoundry", "dir"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, ".cloudfoundry", "other"), []byte("other"), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, ".cloudfoundry", "dir", "other"), []byte("other"), 0644)).To(Succeed())
			})
			It("excludes .cloudfoundry directory", func() {
				Expect(supplier.CalcChecksum()).To(Equal("d8be25466f8d12112d354e1a4add36a3"))
			})
		})
	})

	Describe("InstallGems", func() {
		const windowsWarning = "**WARNING** Windows line endings detected in Gemfile. Your app may fail to stage. Please use UNIX line endings."

		PIt("BACK FILL", func() {})

		Context("Windows Gemfile", func() {
			BeforeEach(func() {
				mockVersions.EXPECT().HasWindowsGemfileLock().Return(false, nil)
				mockVersions.EXPECT().Gemfile().AnyTimes().Return(filepath.Join(buildDir, "Gemfile"))
				mockCommand.EXPECT().Run(gomock.Any()).AnyTimes()
				mockManifest.EXPECT().AllDependencyVersions("bundler").Return([]string{"1.2.3"})
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte("source \"https://rubygems.org\"\r\ngem \"rack\"\r\n"), 0644)).To(Succeed())
			})
			It("Warns the user", func() {
				Expect(supplier.InstallGems()).To(Succeed())
				Expect(buffer.String()).To(ContainSubstring(windowsWarning))
			})
		})

		Context("UNIX Gemfile", func() {
			BeforeEach(func() {
				mockVersions.EXPECT().HasWindowsGemfileLock().Return(false, nil)
				mockVersions.EXPECT().Gemfile().AnyTimes().Return(filepath.Join(buildDir, "Gemfile"))
				mockCommand.EXPECT().Run(gomock.Any()).AnyTimes()
				mockManifest.EXPECT().AllDependencyVersions("bundler").Return([]string{"1.2.3"})
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte("source \"https://rubygems.org\"\ngem \"rack\"\n"), 0644)).To(Succeed())
			})
			It("Does not warn the user", func() {
				Expect(supplier.InstallGems()).To(Succeed())
				Expect(buffer.String()).ToNot(ContainSubstring(windowsWarning))
			})
		})

		Context("Windows Gemfile.lock", func() {
			const gemfileLock = "GEM\n  remote: https://rubygems.org/\n  specs:\n    rack (1.5.2)\n\nPLATFORMS\n  x64-mingw32\n\nDEPENDENCIES\n  rack\n"
			const newGemfileLock = "new lockfile"
			BeforeEach(func() {
				mockVersions.EXPECT().HasWindowsGemfileLock().Return(true, nil)
				mockVersions.EXPECT().Gemfile().AnyTimes().Return(filepath.Join(buildDir, "Gemfile"))
				mockManifest.EXPECT().AllDependencyVersions("bundler").Return([]string{"1.2.3"})
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte("source \"https://rubygems.org\"\r\ngem \"rack\"\r\n"), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile.lock"), []byte(gemfileLock), 0644)).To(Succeed())
			})

			It("runs bundler without the Gemfile.lock and copies the Gemfile.lock it creates to the dep directory", func() {
				mockCommand.EXPECT().Run(gomock.Any()).AnyTimes().Do(func(cmd *exec.Cmd) {
					if cmd.Args[1] == "install" {
						Expect(cmd.Args).ToNot(ContainElement("--deployment"))
						Expect(filepath.Join(cmd.Dir, "Gemfile")).To(BeAnExistingFile())
						Expect(filepath.Join(cmd.Dir, "Gemfile.lock")).ToNot(BeAnExistingFile())
						Expect(ioutil.WriteFile(filepath.Join(cmd.Dir, "Gemfile.lock"), []byte(newGemfileLock), 0644)).To(Succeed())
					}
				})
				Expect(supplier.InstallGems()).To(Succeed())

				Expect(ioutil.ReadFile(filepath.Join(buildDir, "Gemfile.lock"))).To(ContainSubstring(gemfileLock))
				Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "Gemfile.lock"))).To(ContainSubstring(newGemfileLock))
			})

			It("runs bundler in a copy so it does not change the build directory", func() {
				installCalled := false
				mockCommand.EXPECT().Run(gomock.Any()).AnyTimes().Do(func(cmd *exec.Cmd) {
					if cmd.Args[1] == "install" {
						Expect(cmd.Dir).ToNot(Equal(buildDir))
						installCalled = true
					}
				})
				Expect(supplier.InstallGems()).To(Succeed())
				Expect(installCalled).To(BeTrue())
			})
		})
	})

	Describe("InstallJVM", func() {
		Context("app/.jdk exists", func() {
			BeforeEach(func() {
				Expect(os.Mkdir(filepath.Join(buildDir, ".jdk"), 0755)).To(Succeed())
			})
			It("skips jdk install", func() {
				Expect(supplier.InstallJVM()).To(Succeed())

				Expect(buffer.String()).To(ContainSubstring("Using pre-installed JDK"))
				Expect(filepath.Join(depsDir, depsIdx, "jvm")).ToNot(BeADirectory())
			})
		})

		Context("app/.jdk does not exist", func() {
			BeforeEach(func() {
				mockManifest.EXPECT().InstallOnlyVersion("openjdk1.8-latest", gomock.Any()).Do(func(_, path string) error {
					Expect(os.MkdirAll(filepath.Join(path, "bin"), 0755)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(path, "bin", "java"), []byte("java.exe"), 0755)).To(Succeed())
					return nil
				})
			})

			It("installs and links the JDK", func() {
				Expect(supplier.InstallJVM()).To(Succeed())
				Expect(filepath.Join(depsDir, depsIdx, "jvm", "bin", "java")).To(BeAnExistingFile())
				Expect(filepath.Join(depsDir, depsIdx, "bin", "java")).To(BeAnExistingFile())
			})

			It("writes jruby default env vars to profile.d", func() {
				Expect(supplier.InstallJVM()).To(Succeed())
				body, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "jruby.sh"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(body)).To(ContainSubstring(`export JAVA_MEM=${JAVA_MEM:--Xmx${JVM_MAX_HEAP:-384}m}`))
			})
		})
	})

	Describe("CreateDefaultEnv", func() {
		AfterEach(func() {
			_ = os.Unsetenv("RAILS_ENV")
			_ = os.Unsetenv("RACK_ENV")
			_ = os.Unsetenv("RAILS_GROUPS")
		})

		It("Sets RAILS_ENV", func() {
			Expect(supplier.CreateDefaultEnv()).To(Succeed())
			Expect(os.Getenv("RAILS_ENV")).To(Equal("production"))
		})
		It("Sets RAILS_GROUPS", func() {
			Expect(supplier.CreateDefaultEnv()).To(Succeed())
			Expect(os.Getenv("RAILS_GROUPS")).To(Equal("assets"))
		})
		It("Sets RACK_ENV", func() {
			Expect(supplier.CreateDefaultEnv()).To(Succeed())
			Expect(os.Getenv("RACK_ENV")).To(Equal("production"))
		})
		It("Sets RAILS_ENV in env directory", func() {
			Expect(supplier.CreateDefaultEnv()).To(Succeed())
			data, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "RAILS_ENV"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("production"))
		})
		It("Sets RAILS_GROUPS in env directory", func() {
			Expect(supplier.CreateDefaultEnv()).To(Succeed())
			data, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "RAILS_GROUPS"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("assets"))
		})
		It("Sets RACK_ENV in env directory", func() {
			Expect(supplier.CreateDefaultEnv()).To(Succeed())
			data, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "RACK_ENV"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("production"))
		})

		Context("RAILS_ENV is set", func() {
			BeforeEach(func() { _ = os.Setenv("RAILS_ENV", "test") })

			It("does not set RAILS_ENV", func() {
				Expect(supplier.CreateDefaultEnv()).To(Succeed())
				Expect(os.Getenv("RAILS_ENV")).To(Equal("test"))
			})
			It("does not set RAILS_ENV in env directory", func() {
				Expect(supplier.CreateDefaultEnv()).To(Succeed())
				Expect(filepath.Join(depsDir, depsIdx, "env", "RAILS_ENV")).ToNot(BeAnExistingFile())
			})
		})

		Context("RAILS_GROUPS is set", func() {
			BeforeEach(func() { _ = os.Setenv("RAILS_GROUPS", "test") })

			It("does not set RAILS_ENV", func() {
				Expect(supplier.CreateDefaultEnv()).To(Succeed())
				Expect(os.Getenv("RAILS_GROUPS")).To(Equal("test"))
			})
			It("does not set RAILS_ENV in env directory", func() {
				Expect(supplier.CreateDefaultEnv()).To(Succeed())
				Expect(filepath.Join(depsDir, depsIdx, "env", "RAILS_GROUPS")).ToNot(BeAnExistingFile())
			})
		})

		Context("RACK_ENV is set", func() {
			BeforeEach(func() { _ = os.Setenv("RACK_ENV", "test") })

			It("does not set RACK_ENV", func() {
				Expect(supplier.CreateDefaultEnv()).To(Succeed())
				Expect(os.Getenv("RACK_ENV")).To(Equal("test"))
			})
			It("does not set RACK_ENV in env directory", func() {
				Expect(supplier.CreateDefaultEnv()).To(Succeed())
				Expect(filepath.Join(depsDir, depsIdx, "env", "RACK_ENV")).ToNot(BeAnExistingFile())
			})
		})
	})

	Describe("WriteProfileD", func() {
		BeforeEach(func() {
			mockCommand.EXPECT().Output(buildDir, "node", "--version").AnyTimes().Return("v8.2.1", nil)
		})
		Describe("SecretKeyBase", func() {
			Context("Rails >= 4.1", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().RubyEngineVersion().Return("2.3.19", nil)
					mockVersions.EXPECT().HasGemVersion("rails", ">=4.1.0.beta1").Return(true, nil)
				})

				Context("SECRET_KEY_BASE is cached", func() {
					BeforeEach(func() {
						mockCache.EXPECT().Metadata().Return(&cache.Metadata{SecretKeyBase: "foobar"})
					})
					It("writes the cached SECRET_KEY_BASE to profile.d", func() {
						Expect(supplier.WriteProfileD("enginename")).To(Succeed())
						contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "ruby.sh"))
						Expect(err).ToNot(HaveOccurred())
						Expect(string(contents)).To(ContainSubstring("export SECRET_KEY_BASE=${SECRET_KEY_BASE:-foobar}"))
					})
				})

				Context("SECRET_KEY_BASE is not cached", func() {
					BeforeEach(func() {
						mockCache.EXPECT().Metadata().Return(&cache.Metadata{})
						mockCommand.EXPECT().Output(buildDir, "bundle", "exec", "rake", "secret").Return("abcdef", nil)
					})
					It("writes default SECRET_KEY_BASE to profile.d", func() {
						Expect(supplier.WriteProfileD("enginename")).To(Succeed())
						contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "ruby.sh"))
						Expect(err).ToNot(HaveOccurred())
						Expect(string(contents)).To(ContainSubstring("export SECRET_KEY_BASE=${SECRET_KEY_BASE:-abcdef}"))
					})
				})
			})
			Context("NOT Rails >= 4.1", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().RubyEngineVersion().Return("2.3.19", nil)
					mockVersions.EXPECT().HasGemVersion("rails", ">=4.1.0.beta1").Return(false, nil)
				})
				It("does not set default SECRET_KEY_BASE in profile.d", func() {
					Expect(supplier.WriteProfileD("enginename")).To(Succeed())
					contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "ruby.sh"))
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).ToNot(ContainSubstring("SECRET_KEY_BASE"))
				})
			})
		})

		Describe("Default Rails ENVS", func() {
			BeforeEach(func() {
				mockVersions.EXPECT().RubyEngineVersion().Return("2.3.19", nil)
				mockVersions.EXPECT().HasGemVersion("rails", ">=4.1.0.beta1").Return(false, nil)
			})

			It("writes default RAILS_ENV to profile.d", func() {
				Expect(supplier.WriteProfileD("somerubyengine")).To(Succeed())
				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "ruby.sh"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring("export RAILS_ENV=${RAILS_ENV:-production}"))
			})

			It("writes default RAILS_SERVE_STATIC_FILES to profile.d", func() {
				Expect(supplier.WriteProfileD("somerubyengine")).To(Succeed())
				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "ruby.sh"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring("export RAILS_SERVE_STATIC_FILES=${RAILS_SERVE_STATIC_FILES:-enabled}"))
			})

			It("writes default RAILS_LOG_TO_STDOUT to profile.d", func() {
				Expect(supplier.WriteProfileD("somerubyengine")).To(Succeed())
				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "ruby.sh"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring("export RAILS_LOG_TO_STDOUT=${RAILS_LOG_TO_STDOUT:-enabled}"))
			})

			It("writes default GEM_PATH to profile.d", func() {
				Expect(supplier.WriteProfileD("somerubyengine")).To(Succeed())
				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "ruby.sh"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring("export GEM_PATH=${GEM_PATH:-$DEPS_DIR/9/vendor_bundle/somerubyengine/2.3.19:$DEPS_DIR/9/gem_home:$DEPS_DIR/9/bundler}"))
			})
		})
	})

	Describe("DetermineRuby", func() {
		Context("MRI", func() {
			BeforeEach(func() {
				mockVersions.EXPECT().Engine().Return("ruby", nil)
			})

			Context("version determined from Gemfile", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().Version().Return("2.3.1", nil)
				})

				It("returns the engine and version", func() {
					engine, version, err := supplier.DetermineRuby()
					Expect(err).ToNot(HaveOccurred())
					Expect(engine).To(Equal("ruby"))
					Expect(version).To(Equal("2.3.1"))
				})
			})

			Context("version not determined from Gemfile", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().Version().Return("", nil)
					mockManifest.EXPECT().DefaultVersion("ruby").Return(libbuildpack.Dependency{Version: "9.10.11"}, nil)
				})

				It("returns ruby with the default from the manifest", func() {
					engine, version, err := supplier.DetermineRuby()
					Expect(err).ToNot(HaveOccurred())
					Expect(engine).To(Equal("ruby"))
					Expect(version).To(Equal("9.10.11"))
				})

				It("logs a warning", func() {
					_, _, err := supplier.DetermineRuby()
					Expect(err).ToNot(HaveOccurred())
					Expect(buffer.String()).To(ContainSubstring("You have not declared a Ruby version in your Gemfile."))
					Expect(buffer.String()).To(ContainSubstring("Defaulting to 9.10.11"))
				})
			})

			Context("version in Gemfile not in manifest", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().Version().Return("", errors.New(""))
				})

				It("returns an error", func() {
					_, _, err := supplier.DetermineRuby()
					Expect(err).To(HaveOccurred())
				})
			})

		})
		Context("jruby", func() {
			BeforeEach(func() {
				mockVersions.EXPECT().Engine().Return("jruby", nil)
			})
			Context("version determined from Gemfile", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().JrubyVersion().Return("ruby-3.1.2-jruby-2.1.6", nil)
				})

				It("returns the engine and version", func() {
					engine, version, err := supplier.DetermineRuby()
					Expect(err).ToNot(HaveOccurred())
					Expect(engine).To(Equal("jruby"))
					Expect(version).To(Equal("ruby-3.1.2-jruby-2.1.6"))
				})
			})
			Context("version in Gemfile not in manifest", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().JrubyVersion().Return("", errors.New(""))
				})

				It("returns an error", func() {
					_, _, err := supplier.DetermineRuby()
					Expect(err).To(HaveOccurred())
				})
			})
		})
		Context("other", func() {
			BeforeEach(func() {
				mockVersions.EXPECT().Engine().Return("rubinius", nil)
			})
			It("returns an error", func() {
				_, _, err := supplier.DetermineRuby()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("InstallYarn", func() {
		Context("app has yarn.lock file", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "yarn.lock"), []byte("contents"), 0644)).To(Succeed())
			})
			It("installs yarn", func() {
				mockManifest.EXPECT().InstallOnlyVersion("yarn", gomock.Any()).Do(func(_, tempDir string) error {
					Expect(os.MkdirAll(filepath.Join(tempDir, "dist", "bin"), 0755)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(tempDir, "dist", "bin", "yarn"), []byte("contents"), 0644)).To(Succeed())
					return nil
				})
				Expect(supplier.InstallYarn()).To(Succeed())

				Expect(filepath.Join(depsDir, depsIdx, "bin", "yarn")).To(BeAnExistingFile())
				data, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "bin", "yarn"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("contents"))
			})
		})
		Context("app does not have a yarn.lock file", func() {
			It("does NOT install yarn", func() {
				Expect(supplier.InstallYarn()).To(Succeed())
				Expect(filepath.Join(depsDir, depsIdx, "bin", "yarn")).ToNot(BeAnExistingFile())
			})
		})
	})

	Describe("NeedsNode", func() {
		Context("node is not already installed", func() {
			BeforeEach(func() {
				mockCommand.EXPECT().Output(buildDir, "node", "--version").AnyTimes().Return("", fmt.Errorf("could not find node"))
			})
			Context("webpacker is installed", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().HasGemVersion("webpacker", ">=0.0.0").Return(true, nil)
					mockVersions.EXPECT().HasGemVersion(gomock.Any(), ">=0.0.0").AnyTimes().Return(false, nil)
				})
				It("returns true", func() {
					Expect(supplier.NeedsNode()).To(BeTrue())
				})
			})
			Context("execjs is installed", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().HasGemVersion("execjs", ">=0.0.0").Return(true, nil)
					mockVersions.EXPECT().HasGemVersion(gomock.Any(), ">=0.0.0").AnyTimes().Return(false, nil)
				})
				It("returns true", func() {
					Expect(supplier.NeedsNode()).To(BeTrue())
				})
			})
			Context("neither webpacker nor execjs are installed", func() {
				BeforeEach(func() {
					mockVersions.EXPECT().HasGemVersion(gomock.Any(), ">=0.0.0").AnyTimes().Return(false, nil)
				})
				It("returns false", func() {
					Expect(supplier.NeedsNode()).To(BeFalse())
				})
			})
		})
		Context("node is already installed", func() {
			BeforeEach(func() {
				mockCommand.EXPECT().Output(buildDir, "node", "--version").AnyTimes().Return("v8.2.1", nil)
			})
			It("returns false", func() {
				Expect(supplier.NeedsNode()).To(BeFalse())
			})
			It("informs the user that node is being skipped", func() {
				supplier.NeedsNode()
				Expect(buffer.String()).To(ContainSubstring("Skipping install of nodejs since it has been supplied"))
			})
		})
	})
})
