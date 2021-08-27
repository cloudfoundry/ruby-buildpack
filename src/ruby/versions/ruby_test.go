package versions_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/ruby-buildpack/src/ruby/versions"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=ruby.go --destination=mocks_ruby_test.go --package=versions_test

var _ = Describe("Ruby", func() {
	var (
		mockCtrl     *gomock.Controller
		mockManifest *MockManifest
		tmpDir       string
		depDir       string
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockManifest = NewMockManifest(mockCtrl)
		mockManifest.EXPECT().AllDependencyVersions("bundler").Return([]string{"1.17.2"}).AnyTimes()

		var err error
		tmpDir, err = ioutil.TempDir("", "versions.ruby")
		Expect(err).ToNot(HaveOccurred())
		depDir, err = ioutil.TempDir("", "tmpDepdir")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("HasWindowsGemfileLock", func() {
		Context("Gemfile.lock has only mingw/mswin platforms", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(windowsOnlyGemfileLockFixture), 0644)).To(Succeed())
			})

			It("returns true", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				result, err := v.HasWindowsGemfileLock()
				Expect(err).To(BeNil())
				Expect(result).To(BeTrue())
			})
		})

		Context("Gemfile.lock has mingw and ruby platforms and windows line endings", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(windowsEndingsGemfileLockFixture), 0644)).To(Succeed())
			})

			It("returns true", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				result, err := v.HasWindowsGemfileLock()
				Expect(err).To(BeNil())
				Expect(result).To(BeTrue())
			})
		})

		Context("Gemfile.lock has mingw platform and a ruby platform", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(bothGemfileLockFixture), 0644)).To(Succeed())
			})

			It("returns false", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				result, err := v.HasWindowsGemfileLock()
				Expect(err).To(BeNil())
				Expect(result).To(BeFalse())
			})
		})

		Context("Gemfile.lock has ruby platform", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(rubyGemfileLockFixture), 0644)).To(Succeed())
			})

			It("returns false", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				result, err := v.HasWindowsGemfileLock()
				Expect(err).To(BeNil())
				Expect(result).To(BeFalse())
			})
		})

		Context("Gemfile.lock has jruby platform", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(jrubyGemfileLockFixture), 0644)).To(Succeed())
			})

			It("returns false", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				result, err := v.HasWindowsGemfileLock()
				Expect(err).To(BeNil())
				Expect(result).To(BeFalse())
			})
		})

		Context("Gemfile.lock has linux platform", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(linuxGemfileLockFixture), 0644)).To(Succeed())
			})

			It("returns false", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				result, err := v.HasWindowsGemfileLock()
				Expect(err).To(BeNil())
				Expect(result).To(BeFalse())
			})
		})

		Context("Gemfile.lock has mingw platform and linux platform", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(mingwLinuxGemfileLockFixture), 0644)).To(Succeed())
			})

			It("returns false", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				result, err := v.HasWindowsGemfileLock()
				Expect(err).To(BeNil())
				Expect(result).To(BeFalse())
			})
		})

		Context("Gemfile.lock does not exist", func() {
			It("returns false", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.HasWindowsGemfileLock()).To(BeFalse())
			})
		})
	})

	Describe("Engine", func() {
		Context("Gemfile has a mri", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
			})

			It("returns ruby", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.Engine()).To(Equal("ruby"))
			})
		})

		Context("Gemfile has jruby", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby '2.2.3', :engine => 'jruby', :engine_version => '9.1.12.0'`), 0644)).To(Succeed())
			})

			It("returns jruby", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.Engine()).To(Equal("jruby"))
			})
		})

		Context("Gemfile has no constraint", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(``), 0644)).To(Succeed())
			})

			It("returns ruby", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.Engine()).To(Equal("ruby"))
			})
		})

		Context("gemfile doesn't exist", func() {
			It("returns ruby", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.Engine()).To(Equal("ruby"))
			})
		})

		Context("Gemfile does a puts", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte("ruby '2.2.3'\nputs 'Hello'\nSTDERR.puts 'Bye'\n"), 0644)).To(Succeed())
			})

			It("stdout from gemfile does not create problems", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.Engine()).To(Equal("ruby"))
			})
		})
	})

	Describe("Version", func() {
		Context("Gemfile has a constraint", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
			})

			It("returns highest matching version", func() {
				mockManifest.EXPECT().AllDependencyVersions("ruby").Return([]string{"1.2.3", "2.2.3", "2.2.4", "2.2.1", "2.3.3", "3.1.2"})
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.Version()).To(Equal("2.2.4"))
			})

			It("errors if no matching versions", func() {
				mockManifest.EXPECT().AllDependencyVersions("ruby").Return([]string{"1.2.3", "3.1.2"})
				v := versions.New(tmpDir, depDir, mockManifest)
				_, err := v.Version()
				Expect(err).To(MatchError("Running ruby: No Matching versions, ruby ~> 2.2.0 not found in this buildpack"))
			})
		})

		Context("Gemfile has no constraint", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(``), 0644)).To(Succeed())
			})

			It("returns the default version from the manifest", func() {
				mockManifest.EXPECT().AllDependencyVersions("ruby").Return([]string{"1.2.3", "2.2.3", "2.2.4", "2.2.1", "3.1.2"})
				v := versions.New(tmpDir, depDir, mockManifest)
				version, err := v.Version()
				Expect(err).NotTo(HaveOccurred())
				Expect(version).To(Equal(""))
			})
		})

		Context("BUNDLE_GEMFILE env var is set", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile-App"), []byte(`ruby "~>2.3.0"`), 0644)).To(Succeed())
				os.Setenv("BUNDLE_GEMFILE", "Gemfile-App")
			})
			AfterEach(func() { os.Unsetenv("BUNDLE_GEMFILE") })

			It("returns highest matching version", func() {
				mockManifest.EXPECT().AllDependencyVersions("ruby").Return([]string{"1.2.3", "2.2.3", "2.2.4", "2.2.1", "2.3.3", "3.1.2"})
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.Version()).To(Equal("2.3.3"))
			})

			It("errors if no matching versions", func() {
				mockManifest.EXPECT().AllDependencyVersions("ruby").Return([]string{"1.2.3", "2.2.0", "3.1.2"})
				v := versions.New(tmpDir, depDir, mockManifest)
				_, err := v.Version()
				Expect(err).To(MatchError("Running ruby: No Matching versions, ruby ~> 2.3.0 not found in this buildpack"))
			})
		})
	})

	Describe("JrubyVersion", func() {
		Context("Gemfile has a constraint", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby '2.3.3', :engine => 'jruby', :engine_version => '9.1.12.0'`), 0644)).To(Succeed())
			})
			It("returns the requested version", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.JrubyVersion()).To(Equal("9.1.12.0"))
			})
		})

		Context("BUNDLE_GEMFILE env var is set", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby '2.3.3', :engine => 'jruby', :engine_version => '9.1.12.0'`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile-App"), []byte(`ruby '2.4.4', :engine => 'jruby', :engine_version => '9.2.13.0'`), 0644)).To(Succeed())
				os.Setenv("BUNDLE_GEMFILE", "Gemfile-App")
			})
			AfterEach(func() { os.Unsetenv("BUNDLE_GEMFILE") })
			It("returns the requested version", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.JrubyVersion()).To(Equal("9.2.13.0"))
			})
		})
	})

	Describe("RubyEngineVersion", func() {
		It("returns the gem simplified ruby version", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			version, err := v.RubyEngineVersion()
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(MatchRegexp("^\\d+\\.\\d+.0$"))
		})
	})

	Describe("HasGem", func() {
		BeforeEach(func() {
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`gem 'roda'`), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(`GEM
  specs:
    rack (2.0.3)
    roda (2.28.0)
      rack

PLATFORMS
  ruby

DEPENDENCIES
  roda
`), 0644)).To(Succeed())
		})

		It("returns true for roda", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			found, err := v.HasGem("roda")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
		})

		It("returns false for rails", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			Expect(v.HasGem("rails")).To(BeFalse())
		})
	})

	Describe("GemMajorVersion", func() {
		BeforeEach(func() {
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`gem 'roda'`), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(`GEM
  specs:
    rack (2.0.3)
    roda (4.28.0.beta1)
      rack

PLATFORMS
  ruby

DEPENDENCIES
  roda
`), 0644)).To(Succeed())
		})

		It("returns 2 for rack", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			Expect(v.GemMajorVersion("rack")).To(Equal(2))
		})

		It("returns 4 for roda", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			Expect(v.GemMajorVersion("roda")).To(Equal(4))
		})

		It("returns -1 for rails", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			Expect(v.GemMajorVersion("rails")).To(Equal(-1))
		})
	})

	Describe("HasGemVersion", func() {
		BeforeEach(func() {
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`gem 'roda'`), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(`GEM
  specs:
    rack (2.0.3)
    roda (2.28.0)
      rack

PLATFORMS
  ruby

DEPENDENCIES
  roda
`), 0644)).To(Succeed())
		})

		It("returns true for >=2.28.0 for roda", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			match, err := v.HasGemVersion("roda", ">=2.28.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(match).To(BeTrue())
		})

		It("returns false for <2.28.0 for roda", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			match, err := v.HasGemVersion("roda", "<2.28.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(match).To(BeFalse())
		})

		It("returns true for >=2.2.0, <=3.0.0 for roda", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			match, err := v.HasGemVersion("roda", ">=2.2.0", "<=3.0.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(match).To(BeTrue())
		})

		It("returns false for >=2.2.0, <=2.3.0 for roda", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			match, err := v.HasGemVersion("roda", ">=2.2.0", "<=2.3.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(match).To(BeFalse())
		})

		It("returns false for rails", func() {
			v := versions.New(tmpDir, depDir, mockManifest)
			match, err := v.HasGemVersion("rails", "1.0.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(match).To(BeFalse())
		})
	})

	Describe("VersionConstraint", func() {
		var v *versions.Versions
		BeforeEach(func() {
			v = versions.New(tmpDir, depDir, mockManifest)
		})

		It("returns true for 2.28.1 >=2.28.0", func() {
			match, err := v.VersionConstraint("2.28.1", ">=2.28.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(match).To(BeTrue())
		})

		It("returns true for 3.0.1.2 >=2.28.0", func() {
			match, err := v.VersionConstraint("3.0.1.2", ">=2.28.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(match).To(BeTrue())
		})

		It("returns false for 1.50.12 >=2.28.0", func() {
			match, err := v.VersionConstraint("1.50.12", ">=2.28.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(match).To(BeFalse())
		})
	})

	Describe("BundledWithVersion", func() {
		Context("Gemfile.lock has a Bundled With version", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(`
PLATFORMS
   ruby

BUNDLED WITH
   2.1.3`), 0644)).To(Succeed())
			})

			It("returns the requested version", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.BundledWithVersion()).To(Equal("2.1.3"))
			})
		})

		Context("Gemfile.lock doesn't have a Bundled With version", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile.lock"), []byte(`
PLATFORMS
   ruby
`), 0644)).To(Succeed())
			})

			It("returns the requested version", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.BundledWithVersion()).To(Equal(""))
			})
		})

		Context("There is no Gemfile.lock", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(`ruby "~>2.2.0"`), 0644)).To(Succeed())
			})

			It("returns the requested version", func() {
				v := versions.New(tmpDir, depDir, mockManifest)
				Expect(v.BundledWithVersion()).To(Equal(""))
			})
		})
	})
})
