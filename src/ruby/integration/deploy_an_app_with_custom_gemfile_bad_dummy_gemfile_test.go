package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with custom Gemfile and bad file named 'Gemfile'", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(Fixtures("custom_gemfile_bad_dummy_gemfile"))
	})

	It("uses the version of ruby specified in Gemfile-APP", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby 3.1"))
		Expect(app.Stdout.String()).To(ContainSubstring("Installing sinatra 2.2.0"))
	})
})
