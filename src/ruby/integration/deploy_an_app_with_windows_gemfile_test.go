package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with windows Gemfile", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	Context("with windows line endings", func() {
		BeforeEach(func() {
			app = cutlass.New(Fixtures("windows-with-windows-lineendings"))
			app.SetEnv("BP_DEBUG", "1")
		})

		It("warns the user about Windows line endings for Gemfile and removes Gemfile.lock", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("Windows line endings detected in Gemfile. Your app may fail to stage. Please use UNIX line endings."))
			Expect(app.Stdout.String()).To(ContainSubstring("Removing `Gemfile.lock` because it was generated on Windows."))
			Expect(app).To(HaveUnchangedAppdir("BuildDir Checksum Before Supply", "BuildDir Checksum After Supply"))
		})
	})

	Context("with windows as the only platform", func() {
		BeforeEach(func() {
			app = cutlass.New(Fixtures("windows-only"))
			app.SetEnv("BP_DEBUG", "1")
		})

		It("warns the user about Gemfile.lock generated on Windows and removes Gemfile.lock", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("Removing `Gemfile.lock` because it was generated on Windows."))
			Expect(app).To(HaveUnchangedAppdir("BuildDir Checksum Before Supply", "BuildDir Checksum After Supply"))
		})
	})
	Context("with linux line endings and ruby as an additional platform", func() {
		BeforeEach(func() {
			SkipUnlessUncached()
			app = cutlass.New(Fixtures("windows-with-linux-lineendings"))
		})

		It("does not remove Gemfile.lock", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
			Expect(app.Stdout.String()).ToNot(ContainSubstring("Windows line endings detected in Gemfile. Your app may fail to stage. Please use UNIX line endings."))
			Expect(app.Stdout.String()).ToNot(ContainSubstring("Removing `Gemfile.lock` because it was generated on Windows."))
		})
	})
})
