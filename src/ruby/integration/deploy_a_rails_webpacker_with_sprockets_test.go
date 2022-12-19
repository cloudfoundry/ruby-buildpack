package integration_test

import (
	"os/exec"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("pushing a rails webpacker app with sprockets", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(Fixtures("rails6"))
		app.Disk = "1G"
		app.SetEnv("BP_DEBUG", "1")
	})

	It("compiles assets with webpacker", func() {
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		Expect(app.Stdout.String()).To(ContainSubstring("Installing node"))
		Expect(app.Stdout.String()).To(ContainSubstring("Asset precompilation completed"))
		Eventually(app.Stdout.String()).Should(ContainSubstring("Cleaning assets"))

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
