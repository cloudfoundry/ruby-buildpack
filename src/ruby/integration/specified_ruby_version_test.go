package integration_test

import (
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ruby Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "specified_ruby_version"))
	})

	It("", func() {
		PushAppAndConfirm(app)

		By("uses the specified ruby version", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby 2.2.8"))
		})

		By("running a task", func() {
			if !ApiHasTask() {
				Skip("Running against CF without run task support")
			}

			By("can find the specifed ruby in the container", func() {
				_, err := app.RunTask(`echo "RUNNING A TASK: $(ruby --version)"`)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() string { return app.Stdout.String() }, 10*time.Second).Should(ContainSubstring("RUNNING A TASK: ruby 2.2.8"))
			})
		})
	})
})
