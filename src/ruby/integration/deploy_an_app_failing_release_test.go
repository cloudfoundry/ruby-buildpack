package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App which will fail generating release YAML (because tmp isn't a directory)", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "fail_release"))
	})

	Context("Single/Final buildpack", func() {
		BeforeEach(func() {
			app.Buildpacks = []string{"ruby_buildpack"}
		})
		It("fails in finalize", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
			Eventually(app.Stdout.String).Should(ContainSubstring("Error writing release YAML"))
		})
	})
})
