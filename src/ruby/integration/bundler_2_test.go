package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ruby Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	It("uses bundler 2", func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "bundler_2"))
		app.SetEnv("BP_DEBUG", "1")
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Using bundler 2"))
	})

	It("does not use bundler 2 when ruby < 2.3", func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "ruby_2.2_bundler_2"))
		app.SetEnv("BP_DEBUG", "1")
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Using bundler 1"))
	})
})
