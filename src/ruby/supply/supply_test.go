package supply_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	reflect "reflect"
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

type MacTempDir struct{}

func (t *MacTempDir) CopyDirToTemp(dir string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "supply-tests")
	Expect(err).To(BeNil())
	tmpDir = filepath.Join(tmpDir, filepath.Base(dir))
	os.MkdirAll(tmpDir, 0700)
	libbuildpack.CopyDirectory(dir, tmpDir)
	return tmpDir, nil
}

var _ = Describe("Supply", func() {
	var (
		err           error
		buildDir      string
		depsDir       string
		depsIdx       string
		supplier      *supply.Supplier
		logger        *libbuildpack.Logger
		buffer        *bytes.Buffer
		mockCtrl      *gomock.Controller
		mockManifest  *MockManifest
		mockInstaller *MockInstaller
		mockVersions  *MockVersions
		mockCommand   *MockCommand
		mockCache     *MockCache
		mockTempDir   *MacTempDir
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
		mockInstaller = NewMockInstaller(mockCtrl)
		mockVersions = NewMockVersions(mockCtrl)
		mockVersions.EXPECT().Gemfile().AnyTimes().Return(filepath.Join(buildDir, "Gemfile"))
		mockCommand = NewMockCommand(mockCtrl)
		mockCache = NewMockCache(mockCtrl)
		mockTempDir = &MacTempDir{}

		args := []string{buildDir, "", depsDir, depsIdx}
		stager := libbuildpack.NewStager(args, logger, &libbuildpack.Manifest{})

		supplier = &supply.Supplier{
			Stager:    stager,
			Manifest:  mockManifest,
			Installer: mockInstaller,
			Log:       logger,
			Versions:  mockVersions,
			Cache:     mockCache,
			Command:   mockCommand,
			TempDir:   mockTempDir,
		}
	})

	JustBeforeEach(func() {
		Expect(supplier.Setup()).To(Succeed())
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

		handleBundleBinstubRegeneration := func(cmd *exec.Cmd) error {
			if len(cmd.Args) > 5 && reflect.DeepEqual(cmd.Args[0:5], []string{"bundle", "binstubs", "bundler", "--force", "--path"}) {
				Expect(cmd.Args[5]).To(HavePrefix(filepath.Join(depsDir, depsIdx)))
				Expect(os.MkdirAll(cmd.Args[5], 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cmd.Args[5], "bundle"), []byte("new bundle binstub"), 0644)).To(Succeed())
			}
			return nil
		}

		itRegeneratesBundleBinstub := func() {
			It("Re-generates the bundler binstub to replace older, rails-generated ones that are incompatible with bundler > 1.16.0", func() {
				Expect(supplier.InstallGems()).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "binstubs", "bundle"))).To(Equal([]byte("new bundle binstub")))
				Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "bin", "bundle"))).To(Equal([]byte("new bundle binstub")))
			})
		}

		Context("Windows Gemfile", func() {
			BeforeEach(func() {
				mockVersions.EXPECT().HasWindowsGemfileLock().Return(false, nil)
				mockCommand.EXPECT().Run(gomock.Any()).AnyTimes().Do(handleBundleBinstubRegeneration)
				mockManifest.EXPECT().AllDependencyVersions("bundler").Return([]string{"1.2.3"})
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte("source \"https://rubygems.org\"\r\ngem \"rack\"\r\n"), 0644)).To(Succeed())
			})

			itRegeneratesBundleBinstub()

			It("Warns the user", func() {
				Expect(supplier.InstallGems()).To(Succeed())
				Expect(buffer.String()).To(ContainSubstring(windowsWarning))
			})
		})

		Context("UNIX Gemfile", func() {
			BeforeEach(func() {
				os.Setenv("BUNDLE_CONFIG", filepath.Join(depsDir, depsIdx, "bundle_config"))
				mockVersions.EXPECT().HasWindowsGemfileLock().Return(false, nil)
				mockCommand.EXPECT().Run(gomock.Any()).AnyTimes().Do(func(cmd *exec.Cmd) error {
					if len(cmd.Args) > 2 && cmd.Args[1] == "install" {
						Expect(os.MkdirAll(filepath.Join(cmd.Dir, ".bundle"), 0755)).To(Succeed())
						Expect(ioutil.WriteFile(filepath.Join(cmd.Dir, ".bundle", "config"), []byte("new bundle config"), 0644)).To(Succeed())
					} else {
						return handleBundleBinstubRegeneration(cmd)
					}

					return nil
				})
				mockManifest.EXPECT().AllDependencyVersions("bundler").Return([]string{"1.2.3"})
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte("source \"https://rubygems.org\"\ngem \"rack\"\n"), 0644)).To(Succeed())
			})
			AfterEach(func() { os.Unsetenv("BUNDLE_CONFIG") })

			itRegeneratesBundleBinstub()
			It("Does not warn the user", func() {
				Expect(supplier.InstallGems()).To(Succeed())
				Expect(buffer.String()).ToNot(ContainSubstring(windowsWarning))
			})
			It("does not change .bundle/config", func() {
				Expect(os.MkdirAll(filepath.Join(buildDir, ".bundle"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, ".bundle", "config"), []byte("orig content"), 0644)).To(Succeed())
				Expect(ioutil.ReadFile(filepath.Join(buildDir, ".bundle", "config"))).To(Equal([]byte("orig content")))

				Expect(supplier.InstallGems()).To(Succeed())

				Expect(ioutil.ReadFile(filepath.Join(buildDir, ".bundle", "config"))).To(Equal([]byte("orig content")))
			})
		})

		Context("Windows Gemfile.lock", func() {
			Context("With Unix Line Endings", func() {
				const gemfileLock = "GEM\n  remote: https://rubygems.org/\n  specs:\n    rack (1.5.2)\n\nPLATFORMS\n  x64-mingw32\n ruby\n\nDEPENDENCIES\n  rack\n"
				const newGemfileLock = "new lockfile"
				BeforeEach(func() {
					mockVersions.EXPECT().HasWindowsGemfileLock().Return(false, nil)
					mockManifest.EXPECT().AllDependencyVersions("bundler").Return([]string{"1.2.3"})
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte("source \"https://rubygems.org\"\ngem \"rack\"\n"), 0644)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile.lock"), []byte(gemfileLock), 0644)).To(Succeed())
				})

				It("runs bundler with existing Gemfile.lock", func() {
					mockCommand.EXPECT().Run(gomock.Any()).AnyTimes().Do(func(cmd *exec.Cmd) {
						if cmd.Args[1] == "install" {
							Expect(filepath.Join(cmd.Dir, "Gemfile")).To(BeAnExistingFile())
							Expect(filepath.Join(cmd.Dir, "Gemfile.lock")).To(BeAnExistingFile())
						} else {
							handleBundleBinstubRegeneration(cmd)
						}
					})
					Expect(supplier.InstallGems()).To(Succeed())

					Expect(ioutil.ReadFile(filepath.Join(buildDir, "Gemfile.lock"))).To(ContainSubstring(gemfileLock))
					Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "Gemfile.lock"))).To(ContainSubstring(gemfileLock))
				})

				It("runs bundler in a copy so it does not change the build directory", func() {
					installCalled := false
					mockCommand.EXPECT().Run(gomock.Any()).AnyTimes().Do(func(cmd *exec.Cmd) {
						if cmd.Args[1] == "install" {
							Expect(cmd.Dir).ToNot(Equal(buildDir))
							installCalled = true
						} else {
							handleBundleBinstubRegeneration(cmd)
						}
					})
					Expect(supplier.InstallGems()).To(Succeed())
					Expect(installCalled).To(BeTrue())
				})
			})

			Context("With Windows Line Endings", func() {
				const gemfileLock = "GEM\n  remote: https://rubygems.org/\n  specs:\n    rack (1.5.2)\n\nPLATFORMS\n  x64-mingw32\n\nDEPENDENCIES\n  rack\n"
				const newGemfileLock = "new lockfile"
				BeforeEach(func() {
					mockVersions.EXPECT().HasWindowsGemfileLock().Return(true, nil)
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
						} else {
							handleBundleBinstubRegeneration(cmd)
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
						} else {
							handleBundleBinstubRegeneration(cmd)
						}
					})
					Expect(supplier.InstallGems()).To(Succeed())
					Expect(installCalled).To(BeTrue())
				})
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
				mockInstaller.EXPECT().InstallOnlyVersion("openjdk1.8-latest", gomock.Any()).Do(func(_, path string) error {
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

	Describe("EnableLDLibraryPathEnv", func() {
		AfterEach(func() {
			Expect(os.Unsetenv("LD_LIBRARY_PATH")).To(Succeed())
		})
		Context("app has ld_library_path directory", func() {
			BeforeEach(func() {
				Expect(os.Mkdir(filepath.Join(buildDir, "ld_library_path"), 0755)).To(Succeed())
			})
			Context("LD_LIBRARY_PATH is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv("LD_LIBRARY_PATH", "prior_ld_path")).To(Succeed())
				})
				It("Sets LD_LIBRARY_PATH", func() {
					Expect(supplier.EnableLDLibraryPathEnv()).To(Succeed())
					Expect(os.Getenv("LD_LIBRARY_PATH")).To(Equal(filepath.Join(buildDir, "ld_library_path") + ":prior_ld_path"))
				})
				It("Writes LD_LIBRARY_PATH env file for later buildpacks", func() {
					Expect(supplier.EnableLDLibraryPathEnv()).To(Succeed())
					Expect(filepath.Join(depsDir, depsIdx, "env", "LD_LIBRARY_PATH")).To(BeAnExistingFile())
					Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "LD_LIBRARY_PATH"))).To(Equal([]byte(filepath.Join(buildDir, "ld_library_path") + ":prior_ld_path")))
				})
				It("Writes LD_LIBRARY_PATH env file as a profile.d script", func() {
					Expect(supplier.EnableLDLibraryPathEnv()).To(Succeed())
					Expect(filepath.Join(depsDir, depsIdx, "profile.d", "app_lib_path.sh")).To(BeAnExistingFile())
					Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "app_lib_path.sh"))).To(Equal([]byte(`export LD_LIBRARY_PATH="$HOME/ld_library_path$([[ ! -z "${LD_LIBRARY_PATH:-}" ]] && echo ":$LD_LIBRARY_PATH")"`)))
				})
			})

			Context("LD_LIBRARY_PATH is NOT set", func() {
				BeforeEach(func() {
					Expect(os.Unsetenv("LD_LIBRARY_PATH")).To(Succeed())
				})
				It("Sets LD_LIBRARY_PATH", func() {
					Expect(supplier.EnableLDLibraryPathEnv()).To(Succeed())
					Expect(os.Getenv("LD_LIBRARY_PATH")).To(Equal(filepath.Join(buildDir, "ld_library_path")))
				})
				It("Writes LD_LIBRARY_PATH env file for later buildpacks", func() {
					Expect(supplier.EnableLDLibraryPathEnv()).To(Succeed())
					Expect(filepath.Join(depsDir, depsIdx, "env", "LD_LIBRARY_PATH")).To(BeAnExistingFile())
					Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "LD_LIBRARY_PATH"))).To(Equal([]byte(filepath.Join(buildDir, "ld_library_path"))))
				})
				It("Writes LD_LIBRARY_PATH env file as a profile.d script", func() {
					Expect(supplier.EnableLDLibraryPathEnv()).To(Succeed())
					Expect(filepath.Join(depsDir, depsIdx, "profile.d", "app_lib_path.sh")).To(BeAnExistingFile())
					Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "profile.d", "app_lib_path.sh"))).To(Equal([]byte(`export LD_LIBRARY_PATH="$HOME/ld_library_path$([[ ! -z "${LD_LIBRARY_PATH:-}" ]] && echo ":$LD_LIBRARY_PATH")"`)))
				})
			})
		})

		Context("app does NOT have ld_library_path directory", func() {
			var oldLibraryPath string
			BeforeEach(func() {
				oldLibraryPath = os.Getenv("LD_LIBRARY_PATH")
				Expect(os.Setenv("LD_LIBRARY_PATH", "/foo/lib")).To(Succeed())
			})

			AfterEach(func() {
				Expect(os.Setenv("LD_LIBRARY_PATH", oldLibraryPath)).To(Succeed())
			})

			It("Does not change LD_LIBRARY_PATH", func() {
				Expect(supplier.EnableLDLibraryPathEnv()).To(Succeed())
				Expect(os.Getenv("LD_LIBRARY_PATH")).To(Equal("/foo/lib"))
			})
			It("Does not write LD_LIBRARY_PATH env file for later buildpacks", func() {
				Expect(supplier.EnableLDLibraryPathEnv()).To(Succeed())
				Expect(filepath.Join(depsDir, depsIdx, "env", "LD_LIBRARY_PATH")).ToNot(BeAnExistingFile())
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
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte{}, 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile.lock"), []byte{}, 0644)).To(Succeed())
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
						mockCommand.EXPECT().Output(buildDir, "bundle", "exec", "rake", "secret").Return("\n\nabcdef\n\n", nil)
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
		BeforeEach(func() {
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile"), []byte{}, 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "Gemfile.lock"), []byte{}, 0644)).To(Succeed())
		})
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
					mockVersions.EXPECT().JrubyVersion().Return("9.2.0.0", nil)
				})

				It("returns the engine and version", func() {
					engine, version, err := supplier.DetermineRuby()
					Expect(err).ToNot(HaveOccurred())
					Expect(engine).To(Equal("jruby"))
					Expect(version).To(Equal("9.2.0.0"))
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
				mockInstaller.EXPECT().InstallOnlyVersion("yarn", gomock.Any()).Do(func(_, tempDir string) error {
					Expect(os.MkdirAll(filepath.Join(tempDir, "yarn-v1.2.3", "bin"), 0755)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(tempDir, "yarn-v1.2.3", "bin", "yarn"), []byte("contents"), 0644)).To(Succeed())
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

	Describe("UpdateRubygems", func() {
		BeforeEach(func() {
			mockManifest.EXPECT().AllDependencyVersions("rubygems").AnyTimes().Return([]string{"2.6.13"})
		})
		Context("gem version is less than 2.6.13", func() {
			BeforeEach(func() {
				mockCommand.EXPECT().Output(gomock.Any(), "gem", "--version").AnyTimes().Return("2.6.12\n", nil)
				mockVersions.EXPECT().VersionConstraint("2.6.12", ">= 2.6.13").AnyTimes().Return(false, nil)
			})

			It("updates rubygems", func() {
				mockVersions.EXPECT().Engine().Return("ruby", nil)
				mockInstaller.EXPECT().InstallDependency(gomock.Any(), gomock.Any()).Do(func(dep libbuildpack.Dependency, _ string) {
					Expect(dep.Name).To(Equal("rubygems"))
					Expect(dep.Version).To(Equal("2.6.13"))
				})
				mockCommand.EXPECT().Output(gomock.Any(), "ruby", "setup.rb")

				Expect(supplier.UpdateRubygems()).To(Succeed())
			})

			Context("jruby", func() {
				It("skips update of rubygems", func() {
					mockVersions.EXPECT().Engine().Return("jruby", nil)
					Expect(supplier.UpdateRubygems()).To(Succeed())
				})
			})
		})
		Context("gem version is equal to 2.6.13", func() {
			BeforeEach(func() {
				mockCommand.EXPECT().Output(gomock.Any(), "gem", "--version").AnyTimes().Return("2.6.13\n", nil)
				mockVersions.EXPECT().VersionConstraint("2.6.13", ">= 2.6.13").AnyTimes().Return(true, nil)
			})

			It("does nothing", func() {
				Expect(supplier.UpdateRubygems()).To(Succeed())
			})
		})
		Context("gem version is greater than to 2.6.13", func() {
			BeforeEach(func() {
				mockCommand.EXPECT().Output(gomock.Any(), "gem", "--version").AnyTimes().Return("2.6.14\n", nil)
				mockVersions.EXPECT().VersionConstraint("2.6.14", ">= 2.6.13").AnyTimes().Return(true, nil)
			})

			It("does nothing", func() {
				Expect(supplier.UpdateRubygems()).To(Succeed())
			})
		})
	})

	Describe("RewriteShebangs", func() {
		var depDir string
		BeforeEach(func() {
			depDir = filepath.Join(depsDir, depsIdx)
			Expect(os.MkdirAll(filepath.Join(depDir, "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(depDir, "bin", "somescript"), []byte("#!/usr/bin/ruby\n\n\n"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(depDir, "bin", "anotherscript"), []byte("#!//bin/ruby\n\n\n"), 0755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(depDir, "bin", "__ruby__"), 0755)).To(Succeed())
			Expect(os.Symlink(filepath.Join(depDir, "bin", "__ruby__"), filepath.Join(depDir, "bin", "__ruby__SYMLINK"))).To(Succeed())
		})
		It("changes them to #!/usr/bin/env ruby", func() {
			Expect(supplier.RewriteShebangs()).To(Succeed())

			fileContents, err := ioutil.ReadFile(filepath.Join(depDir, "bin", "somescript"))
			Expect(err).ToNot(HaveOccurred())

			secondFileContents, err := ioutil.ReadFile(filepath.Join(depDir, "bin", "anotherscript"))
			Expect(err).ToNot(HaveOccurred())

			Expect(string(fileContents)).To(HavePrefix("#!/usr/bin/env ruby"))
			Expect(string(secondFileContents)).To(HavePrefix("#!/usr/bin/env ruby"))
		})
		It(`also finds files in vendor_bundle/ruby/*/bin/*`, func() {
			Expect(os.MkdirAll(filepath.Join(depDir, "vendor_bundle", "ruby", "2.4.0", "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(depDir, "vendor_bundle", "ruby", "2.4.0", "bin", "somescript"), []byte("#!/usr/bin/ruby\n\n\n"), 0755)).To(Succeed())

			Expect(supplier.RewriteShebangs()).To(Succeed())

			fileContents, err := ioutil.ReadFile(filepath.Join(depDir, "vendor_bundle", "ruby", "2.4.0", "bin", "somescript"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(fileContents)).To(HavePrefix("#!/usr/bin/env ruby"))
		})
	})
	Describe("SymlinkBundlerIntoRubygems", func() {
		var depDir string
		BeforeEach(func() {
			depDir = filepath.Join(depsDir, depsIdx)
			mockVersions.EXPECT().RubyEngineVersion().Return("2.3.4", nil)
			mockManifest.EXPECT().AllDependencyVersions("bundler").Return([]string{"1.2.3"})

			Expect(os.MkdirAll(filepath.Join(depDir, "bundler", "gems", "bundler-1.2.3"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(depDir, "bundler", "gems", "bundler-1.2.3", "file"), []byte("my content"), 0644)).To(Succeed())
		})
		It("Creates a symlink from the installed ruby's gem directory to the installed bundler gem", func() {
			Expect(supplier.SymlinkBundlerIntoRubygems()).To(Succeed())

			fileContents, err := ioutil.ReadFile(filepath.Join(depDir, "ruby", "lib", "ruby", "gems", "2.3.4", "gems", "bundler-1.2.3", "file"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(fileContents)).To(HavePrefix("my content"))
		})
	})
})
