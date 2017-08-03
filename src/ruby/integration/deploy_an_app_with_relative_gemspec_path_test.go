package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with relative gemspec path", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "relative_gemspec_path"))
	})

	It("loads the gem with the relative gemspec path", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Using hola 0.0.0 from source at `gems/hola`"))
	})
})
