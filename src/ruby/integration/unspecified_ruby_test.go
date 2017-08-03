package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ruby Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "unspecified_ruby"))
		app.SetEnv("BP_DEBUG", "1")
	})

	defaultVersion := func(name string) string {
		m := &libbuildpack.Manifest{}
		err := (&libbuildpack.YAML{}).Load(filepath.Join(bpDir, "manifest.yml"), m)
		Expect(err).ToNot(HaveOccurred())
		dep, err := m.DefaultVersion(name)
		Expect(err).ToNot(HaveOccurred())
		Expect(dep.Version).ToNot(Equal(""))
		return dep.Version
	}

	It("uses the default ruby version when ruby version is not specified", func() {
		PushAppAndConfirm(app)
		defaultRubyVersion := defaultVersion("ruby")

		Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby %s", defaultRubyVersion))
	})
})
