package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("requiring execjs", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_execjs"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Installing node 4."))
		Expect(app).To(HaveUnchangedAppdir("BuildDir Checksum Before Supply", "BuildDir Checksum After Supply"))

		Expect(app.GetBody("/")).To(ContainSubstring("Successfully required execjs"))
		Expect(app.Stdout.String()).ToNot(ContainSubstring("ExecJS::RuntimeUnavailable"))

		Expect(app.GetBody("/npm")).To(ContainSubstring("Usage: npm <command>"))
	})
})
