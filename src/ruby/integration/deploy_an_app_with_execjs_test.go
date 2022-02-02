package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("requiring execjs", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(Fixtures("with_execjs"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("installs node and execjs", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Installing node"))
		Expect(app).To(HaveUnchangedAppdir("BuildDir Checksum Before Supply", "BuildDir Checksum After Supply"))

		Expect(app.GetBody("/")).To(ContainSubstring("Successfully required execjs"))
		Expect(app.Stdout.String()).ToNot(ContainSubstring("ExecJS::RuntimeUnavailable"))

		Expect(app.GetBody("/npm")).To(ContainSubstring("npm <command>\n\nUsage:"))
	})
})
