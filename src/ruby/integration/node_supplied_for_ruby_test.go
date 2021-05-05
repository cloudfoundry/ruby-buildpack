package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("running supply nodejs buildpack before the ruby buildpack", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("the app is pushed", func() {
		BeforeEach(func() {
			if ok, err := cutlass.ApiGreaterThan("2.65.1"); err != nil || !ok {
				Skip("API version does not have multi-buildpack support")
			}

			app = cutlass.New(Fixtures("rails5"))
			app.Buildpacks = []string{
				"https://github.com/cloudfoundry/nodejs-buildpack#master",
				"ruby_buildpack",
			}
			app.Disk = "1G"
		})

		It("finds the supplied dependency in the runtime container", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(ContainSubstring("Nodejs Buildpack version"))
			Expect(app.Stdout.String()).To(ContainSubstring("Installing node 14."))

			body, err := app.GetBody("/")
			Expect(err).To(BeNil())
			Expect(body).To(ContainSubstring("Ruby version: ruby 2."))
			Expect(body).To(ContainSubstring("Node version: v10."))
			Expect(body).To(ContainSubstring("/home/vcap/deps/0/node"))

			Expect(app.Stdout.String()).To(ContainSubstring("Skipping install of nodejs since it has been supplied"))
		})
	})
})
