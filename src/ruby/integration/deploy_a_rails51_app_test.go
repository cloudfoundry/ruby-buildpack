package integration_test

import (
	"os/exec"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Rails 5.1 (Webpack/Yarn) App", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(Fixtures("rails51"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("Installs node and runs", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Installing node"))

		Expect(app.GetBody("/")).To(ContainSubstring("Hello World"))
		Eventually(app.Stdout.String).Should(ContainSubstring(`Started GET "/" for`))
		Eventually(app.Stdout.String).Should(ContainSubstring("Cleaning assets"))

		By("Make sure supply does not change BuildDir", func() {
			Expect(app).To(HaveUnchangedAppdir("BuildDir Checksum Before Supply", "BuildDir Checksum After Supply"))
		})

		By("Make sure binstubs work", func() {
			command := exec.Command("cf", "ssh", app.Name, "-c", "/tmp/lifecycle/launcher /home/vcap/app 'rails about' ''")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 10, 0.25).Should(gexec.Exit(0))
		})
	})
})
