package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing an app with Rails 7.1", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("a rails 7.1 app is pushed", func() {
		BeforeEach(func() {
			if ok, err := cutlass.ApiGreaterThan("2.65.1"); err != nil || !ok {
				Skip("API version does not have multi-buildpack support")
			}

			app = cutlass.New(Fixtures("rails7"))
			app.Buildpacks = []string{
				"https://github.com/cloudfoundry/nodejs-buildpack#master",
				"ruby_buildpack",
			}
			app.Disk = "1G"
		})

		It("does not fail calling `rake secret` and instead calls rails secret when rails ~> 7.1 is detected", func() {
			PushAppAndConfirm(app)

			// 7.0 has `rake secret` and `rails secret` available.
			// 7.1 has `rails secret` available, `rake secret` was removed
			Expect(app.Stdout.String()).Not(To(ContainSubstring("Don't know how to build task 'secret'")))
			Expect(app.Stdout.String()).Not(To(ContainSubstring(`Unable to write profile.d: Failed to run 'rake secret': exit status 1`)))

			body, err := app.GetBody("/")
			Expect(err).To(BeNil())
		})
	})
})
