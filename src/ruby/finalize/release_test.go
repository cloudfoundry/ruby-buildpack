package finalize_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"ruby/finalize"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=release.go --destination=mocks_release_test.go --package=finalize_test

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
		mockStager   *MockStager
		mockVersions *MockVersions
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
		mockStager = NewMockStager(mockCtrl)
		mockVersions = NewMockVersions(mockCtrl)

		finalizer = &finalize.Finalizer{
			Stager:   mockStager,
			Versions: mockVersions,
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

	Describe("GenerateReleaseYaml", func() {
		var hasRack, hasThin bool
		var railsVersion int
		BeforeEach(func() {
			hasRack = false
			hasThin = false
			railsVersion = 0
		})
		JustBeforeEach(func() {
			mockVersions.EXPECT().HasGem("rack").Return(hasRack, nil)
			mockVersions.EXPECT().HasGem("thin").Return(hasThin, nil)
			mockVersions.EXPECT().HasGemVersion("rails", ">=4.0.0-beta").AnyTimes().Return(railsVersion >= 4, nil)
			mockVersions.EXPECT().HasGemVersion("rails", ">=3.0.0").AnyTimes().Return(railsVersion >= 3, nil)
			mockVersions.EXPECT().HasGemVersion("rails", ">=2.0.0").AnyTimes().Return(railsVersion >= 2, nil)
		})
		Context("Rails 4+", func() {
			BeforeEach(func() {
				railsVersion = 4
			})
			It("generates web, worker, rake and console process types", func() {
				data, err := finalizer.GenerateReleaseYaml()
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal(map[string]map[string]string{
					"default_process_types": map[string]string{
						"rake":    "bundle exec rake",
						"console": "bin/rails console",
						"web":     "bin/rails server -b 0.0.0.0 -p $PORT -e $RAILS_ENV",
						"worker":  "bundle exec rake jobs:work",
					},
				}))
			})
		})

		Context("Rails 3.x", func() {
			BeforeEach(func() {
				railsVersion = 3
			})
			Context("thin is not present", func() {
				BeforeEach(func() {
					hasThin = false
				})
				It("generates web, worker, rake and console process types", func() {
					data, err := finalizer.GenerateReleaseYaml()
					Expect(err).NotTo(HaveOccurred())
					Expect(data).To(Equal(map[string]map[string]string{
						"default_process_types": map[string]string{
							"rake":    "bundle exec rake",
							"console": "bundle exec rails console",
							"web":     "bundle exec rails server -p $PORT",
							"worker":  "bundle exec rake jobs:work",
						},
					}))
				})
			})
			Context("thin is present", func() {
				BeforeEach(func() {
					hasThin = true
				})
				It("generates web, worker, rake and console process types", func() {
					data, err := finalizer.GenerateReleaseYaml()
					Expect(err).NotTo(HaveOccurred())
					Expect(data).To(Equal(map[string]map[string]string{
						"default_process_types": map[string]string{
							"rake":    "bundle exec rake",
							"console": "bundle exec rails console",
							"web":     "bundle exec thin start -R config.ru -e $RAILS_ENV -p $PORT",
							"worker":  "bundle exec rake jobs:work",
						},
					}))
				})
			})
		})
		Context("Rails 2.x", func() {
			BeforeEach(func() {
				railsVersion = 2
			})
			Context("thin is not present", func() {
				BeforeEach(func() {
					hasThin = false
				})
				It("generates web, worker, rake and console process types", func() {
					data, err := finalizer.GenerateReleaseYaml()
					Expect(err).NotTo(HaveOccurred())
					Expect(data).To(Equal(map[string]map[string]string{
						"default_process_types": map[string]string{
							"rake":    "bundle exec rake",
							"console": "bundle exec script/console",
							"web":     "bundle exec ruby script/server -p $PORT",
							"worker":  "bundle exec rake jobs:work",
						},
					}))
				})
			})
			Context("thin is present", func() {
				BeforeEach(func() {
					hasThin = true
				})
				It("generates web, worker, rake and console process types", func() {
					data, err := finalizer.GenerateReleaseYaml()
					Expect(err).NotTo(HaveOccurred())
					Expect(data).To(Equal(map[string]map[string]string{
						"default_process_types": map[string]string{
							"rake":    "bundle exec rake",
							"console": "bundle exec script/console",
							"web":     "bundle exec thin start -e $RAILS_ENV -p $PORT",
							"worker":  "bundle exec rake jobs:work",
						},
					}))
				})
			})
		})
		Context("Rack", func() {
			BeforeEach(func() {
				hasRack = true
			})
			Context("thin is not present", func() {
				BeforeEach(func() {
					hasThin = false
				})
				It("generates web, rake and console process types", func() {
					data, err := finalizer.GenerateReleaseYaml()
					Expect(err).NotTo(HaveOccurred())
					Expect(data).To(Equal(map[string]map[string]string{
						"default_process_types": map[string]string{
							"rake":    "bundle exec rake",
							"console": "bundle exec irb",
							"web":     "bundle exec rackup config.ru -p $PORT",
						},
					}))
				})
			})
			Context("thin is present", func() {
				BeforeEach(func() {
					hasThin = true
				})
				It("generates web, rake and console process types", func() {
					data, err := finalizer.GenerateReleaseYaml()
					Expect(err).NotTo(HaveOccurred())
					Expect(data).To(Equal(map[string]map[string]string{
						"default_process_types": map[string]string{
							"rake":    "bundle exec rake",
							"console": "bundle exec irb",
							"web":     "bundle exec thin start -R config.ru -e $RACK_ENV -p $PORT",
						},
					}))
				})
			})
		})
		Context("Ruby", func() {
			BeforeEach(func() {
				hasRack = false
				hasThin = false
				railsVersion = 0
			})
			It("generates rake and console process types", func() {
				data, err := finalizer.GenerateReleaseYaml()
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal(map[string]map[string]string{
					"default_process_types": map[string]string{
						"rake":    "bundle exec rake",
						"console": "bundle exec irb",
					},
				}))
			})
		})
	})
})
