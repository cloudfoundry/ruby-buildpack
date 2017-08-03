package cache_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"ruby/cache"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=cache.go --destination=mocks_cache_test.go --package=cache_test

var _ = Describe("Cache", func() {
	var (
		err        error
		buildDir   string
		cacheDir   string
		depsDir    string
		depsIdx    string
		logger     *libbuildpack.Logger
		buffer     *bytes.Buffer
		mockCtrl   *gomock.Controller
		mockYaml   *MockYAML
		mockStager *MockStager
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "ruby-buildpack.build.")
		Expect(err).To(BeNil())

		cacheDir, err = ioutil.TempDir("", "ruby-buildpack.cache.")
		Expect(err).To(BeNil())

		depsDir, err = ioutil.TempDir("", "ruby-buildpack.deps.")
		Expect(err).To(BeNil())

		depsIdx = "23"
		Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)).To(Succeed())

		buffer = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockYaml = NewMockYAML(mockCtrl)
		mockStager = NewMockStager(mockCtrl)
		mockStager.EXPECT().BuildDir().AnyTimes().Return(buildDir)
		mockStager.EXPECT().CacheDir().AnyTimes().Return(cacheDir)
		mockStager.EXPECT().DepDir().AnyTimes().Return(filepath.Join(depsDir, depsIdx))
	})

	AfterEach(func() {
		mockCtrl.Finish()
		Expect(os.RemoveAll(buildDir)).To(Succeed())
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
	})

	Describe("New", func() {
		Context("cache/metadata.yml exists", func() {
			BeforeEach(func() {
				mockYaml.EXPECT().Load(filepath.Join(cacheDir, "metadata.yml"), gomock.Any()).Do(func(_ string, val interface{}) error {
					metadata := val.(*cache.Metadata)
					metadata.Stack = "cflinuxfs9"
					metadata.SecretKeyBase = "abcdef"
					return nil
				})
			})

			It("loads the metadata", func() {
				c, err := cache.New(mockStager, logger, mockYaml)
				Expect(err).ToNot(HaveOccurred())

				Expect(c.Metadata().Stack).To(Equal("cflinuxfs9"))
				Expect(c.Metadata().SecretKeyBase).To(Equal("abcdef"))
			})
		})

		Context("cache/metadata.yml does NOT exist", func() {
			BeforeEach(func() {
				mockYaml.EXPECT().Load(filepath.Join(cacheDir, "metadata.yml"), gomock.Any()).Return(os.ErrNotExist)
			})

			It("initializes the metadata", func() {
				c, err := cache.New(mockStager, logger, mockYaml)
				Expect(err).ToNot(HaveOccurred())

				Expect(c.Metadata().Stack).To(Equal(""))
				Expect(c.Metadata().SecretKeyBase).To(Equal(""))
			})
		})
	})

	Describe("Save", func() {
		var c *cache.Cache
		BeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "vendor_bundle", "adir", "bdir"), 0755)).To(Succeed())
			mockYaml.EXPECT().Load(filepath.Join(cacheDir, "metadata.yml"), gomock.Any()).Return(os.ErrNotExist)
			var err error
			c, err = cache.New(mockStager, logger, mockYaml)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Copies vendor_bundle to cacheDir", func() {
			mockYaml.EXPECT().Write(filepath.Join(cacheDir, "metadata.yml"), gomock.Any()).AnyTimes().Return(nil)
			Expect(c.Save()).To(Succeed())

			Expect(filepath.Join(cacheDir, "vendor_bundle", "adir", "bdir")).To(BeADirectory())
		})

		It("Stores metadata", func() {
			mockYaml.EXPECT().Write(filepath.Join(cacheDir, "metadata.yml"), gomock.Any()).Return(nil)

			Expect(c.Save()).To(Succeed())
		})
	})

	Describe("Restore", func() {
		var c *cache.Cache
		BeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(cacheDir, "vendor_bundle", "adir", "bdir"), 0755)).To(Succeed())
			mockYaml.EXPECT().Load(filepath.Join(cacheDir, "metadata.yml"), gomock.Any()).Do(func(_ string, val interface{}) error {
				metadata := val.(*cache.Metadata)
				metadata.Stack = "cflinuxfs8"
				metadata.SecretKeyBase = "abcdef"
				return nil
			})
			var err error
			c, err = cache.New(mockStager, logger, mockYaml)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("stack matches", func() {
			BeforeEach(func() {
				os.Setenv("CF_STACK", "cflinuxfs8")
			})
			It("restores vendor_bundle directory", func() {
				Expect(c.Restore()).To(Succeed())

				Expect(filepath.Join(depsDir, depsIdx, "vendor_bundle", "adir", "bdir")).To(BeADirectory())
				Expect(filepath.Join(cacheDir, "vendor_bundle")).ToNot(BeADirectory())
			})
		})

		Context("stack differs", func() {
			BeforeEach(func() {
				os.Setenv("CF_STACK", "cflinuxfs9")
			})
			It("does not restore vendor_bundle directory", func() {
				Expect(c.Restore()).To(Succeed())

				Expect(filepath.Join(depsDir, depsIdx, "vendor_bundle")).ToNot(BeADirectory())
				Expect(filepath.Join(cacheDir, "vendor_bundle")).ToNot(BeADirectory())
			})
		})
	})
})
