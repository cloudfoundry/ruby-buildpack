package integration_test

import (
	"path/filepath"

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

	It("uses the default ruby version when ruby version is not specified", func() {
		PushAppAndConfirm(app)
		defaultRubyVersion := DefaultVersion("ruby")

		Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby %s", defaultRubyVersion))
	})
})
