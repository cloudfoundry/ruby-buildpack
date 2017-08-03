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

		Expect(app.GetBody("/")).To(ContainSubstring("Successfully required execjs"))
		Expect(app.Stdout.String()).ToNot(ContainSubstring("ExecJS::RuntimeUnavailable"))

		Expect(app.GetBody("/npm")).To(ContainSubstring("Usage: npm <command>"))

		// Make sure supply does not change BuildDir
		Expect(app.Stdout.String()).To(ContainSubstring("BuildDir Checksum Before Supply: b3d19453a33206783c48720e172bf019"))
		Expect(app.Stdout.String()).To(ContainSubstring("BuildDir Checksum After Supply: b3d19453a33206783c48720e172bf019"))
	})
})
