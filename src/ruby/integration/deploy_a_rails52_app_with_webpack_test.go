package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Rails 5.2 (Webpack/Yarn) App", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails52_webpack"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("Installs node6 and runs", func() {
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("Hello, Rails!"))
		Eventually(app.Stdout.String).Should(ContainSubstring(`Started GET "/" for`))

		Eventually(app.Stdout.String).ShouldNot(ContainSubstring("assets:clean"))

	})
})
