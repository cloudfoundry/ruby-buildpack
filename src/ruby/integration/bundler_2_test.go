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

	It("uses bundler 2", func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "bundler_2"))
		app.SetEnv("BP_DEBUG", "1")
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Using bundler 2"))
	})
})
