package brats_test

import (
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
)

var _ = Describe("Ruby buildpack", func() {
	bratshelper.UnbuiltBuildpack("ruby", CopyBrats)
	bratshelper.DeployingAnAppWithAnUpdatedVersionOfTheSameBuildpack(CopyBrats)
	bratshelper.StagingWithBuildpackThatSetsEOL("ruby", func(_ string) *cutlass.App {
		return CopyBrats("2.6.x")
	})
	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(CopyBrats)
	bratshelper.DeployAppWithExecutableProfileScript("ruby", CopyBrats)
	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)
	bratshelper.ForAllSupportedVersions("ruby", CopyBrats, func(rubyVersion string, app *cutlass.App) {
		PushApp(app)

		By("installs the correct version of Ruby", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby " + rubyVersion))
			Expect(app.GetBody("/version")).To(ContainSubstring(rubyVersion))
		})
		By("runs a simple webserver", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
		By("parses XML with nokogiri", func() {
			Expect(app.GetBody("/nokogiri")).To(ContainSubstring("Hello, World"))
		})
		By("supports EventMachine", func() {
			Expect(app.GetBody("/em")).To(ContainSubstring("Hello, EventMachine"))
		})
		By("encrypts with bcrypt", func() {
			hashedPassword, err := app.GetBody("/bcrypt")
			Expect(err).ToNot(HaveOccurred())
			Expect(bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("Hello, bcrypt"))).ToNot(HaveOccurred())
		})
		By("supports bson", func() {
			Expect(app.GetBody("/bson")).To(ContainSubstring("\x00\x04\x00\x00"))
		})
		By("supports postgres", func() {
			Expect(app.GetBody("/pg")).To(ContainSubstring("could not connect to server: No such file or directory"))
		})
		By("supports mysql2", func() {
			Expect(app.GetBody("/mysql2")).To(ContainSubstring("Unknown MySQL server host 'testing'"))
		})
	})

	bratshelper.ForAllSupportedVersions("jruby", CopyBratsJRuby, func(jrubyVersion string, app *cutlass.App) {
		app.Memory = "2G"
		app.Disk = "1G" //This failed at 300M
		app.StartCommand = "ruby app.rb -p $PORT"
		PushApp(app)

		By("installs the correct version of JRuby", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing jruby " + jrubyVersion))
		})
		By("runs a simple webserver", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello, World"))
		})
		By("parses XML with nokogiri", func() {
			Expect(app.GetBody("/nokogiri")).To(ContainSubstring("Hello, World"))
		})
		By("supports EventMachine", func() {
			Expect(app.GetBody("/em")).To(ContainSubstring("Hello, EventMachine"))
		})
		By("encrypts with bcrypt", func() {
			hashedPassword, err := app.GetBody("/bcrypt")
			Expect(err).ToNot(HaveOccurred())
			Expect(bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("Hello, bcrypt"))).ToNot(HaveOccurred())
		})
		By("supports postgres", func() {
			Expect(app.GetBody("/pg")).To(ContainSubstring("The connection attempt failed."))
		})
		By("supports mysql", func() {
			Expect(app.GetBody("/mysql")).To(ContainSubstring("Communications link failure"))
		})
	})
})
