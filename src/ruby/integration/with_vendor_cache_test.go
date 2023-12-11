package integration_test

import (
	"bytes"
	"github.com/paketo-buildpacks/packit/pexec"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with dependencies installed in vendor/cache", func() {
	var (
		app        *cutlass.App
		cf         pexec.Executable
		config     *os.File
		org, space string
	)

	BeforeEach(func() {
		SkipUnlessCached()

		app = cutlass.New(Fixtures("with_vendor_cache"))

		cf = pexec.NewExecutable("cf")

		var err error
		config, err = ioutil.TempFile("", "security-group")
		Expect(err).NotTo(HaveOccurred())
		defer config.Close()

		_, err = config.WriteString(`[
			{
				"destination": "10.0.0.0-10.255.255.255",
				"ports": "443",
				"protocol": "tcp"
			},
			{
				"destination": "172.16.0.0-172.31.255.255",
				"ports": "443",
				"protocol": "tcp"
			},
			{
				"destination": "192.168.0.0-192.168.255.255",
				"ports": "443",
				"protocol": "tcp"
			}
		]`)
		Expect(err).NotTo(HaveOccurred())

		buffer := bytes.NewBuffer(nil)
		err = cf.Execute(pexec.Execution{
			Args:   []string{"target"},
			Stdout: buffer,
		})
		Expect(err).NotTo(HaveOccurred())

		lines := strings.Split(buffer.String(), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "org:") {
				org = strings.TrimSpace(strings.TrimPrefix(line, "org:"))
			}

			if strings.HasPrefix(line, "space:") {
				space = strings.TrimSpace(strings.TrimPrefix(line, "space:"))
			}
		}

		err = cf.Execute(pexec.Execution{
			Args: []string{"create-space", "offline"},
		})
		Expect(err).NotTo(HaveOccurred())

		err = cf.Execute(pexec.Execution{
			Args: []string{"target", "-o", org, "-s", "offline"},
		})
		Expect(err).NotTo(HaveOccurred())

		err = cf.Execute(pexec.Execution{
			Args: []string{"create-security-group", "offline", config.Name()},
		})
		Expect(err).NotTo(HaveOccurred())

		err = cf.Execute(pexec.Execution{
			Args: []string{"bind-security-group", "offline", org, "offline", "--lifecycle", "staging"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		app = DestroyApp(app)

		err := cf.Execute(pexec.Execution{
			Args: []string{"target", "-o", org, "-s", space},
		})
		Expect(err).NotTo(HaveOccurred())

		err = cf.Execute(pexec.Execution{
			Args: []string{"delete-space", "offline", "-f"},
		})
		Expect(err).NotTo(HaveOccurred())

		err = cf.Execute(pexec.Execution{
			Args: []string{"delete-security-group", "offline", "-f"},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Remove(config.Name())).To(Succeed())
	})

	It("deploys successfully in an offline environment", func() {
		PushAppAndConfirm(app)

		By("has required gems at runtime", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Healthy"))
			Eventually(app.Stdout.String).Should(ContainSubstring("This is red"))
			Eventually(app.Stdout.String).Should(ContainSubstring("This is blue"))
		})
	})
})
