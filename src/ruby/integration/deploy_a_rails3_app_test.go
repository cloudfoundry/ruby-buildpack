package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rails 3 App", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	It("in an online environment", func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails3_mri_200"))
		app.SetEnv("DATABASE_URL", "sqlite3://db/test.db")
		PushAppAndConfirm(app)

		By("the app can be visited in the browser", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("hello"))
		})

		By("the app did not include the static asset or logging gems from Heroku", func() {
			By("the rails 3 plugins are installed automatically", func() {
				files, err := app.Files("/app/vendor/plugins")
				Expect(err).ToNot(HaveOccurred())
				Expect(files).To(ContainElement("/app/vendor/plugins/rails3_serve_static_assets/init.rb"))
				Expect(files).To(ContainElement("/app/vendor/plugins/rails_log_stdout/init.rb"))
			})
		})

		By("we include a rails logger message in the initializer", func() {
			By("the log message is visible in the cf cli app logging", func() {
				Expect(app.Stdout.String()).To(ContainSubstring("Logging is being redirected to STDOUT with rails_log_stdout plugin"))
			})
		})

		By("we include a static asset", func() {
			By("app serves the static asset", func() {
				Expect(app.GetBody("/assets/application.css")).To(ContainSubstring("body{color:red}"))
			})
		})
	})

	AssertNoInternetTraffic("rails3_mri_200")
})
