package integration_test

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"regexp"
	"io/ioutil"
)

var _ = Describe("JRuby App", func() {
	var app *cutlass.App

	AfterEach(func() { app = DestroyApp(app) })

	Context("without start command", func() {
		var dir, rubyVersion, jrubyVersion string

		BeforeEach(func() {
			dir = filepath.Join(bpDir, "fixtures", "sinatra_jruby")
			data, err:= ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
			Expect(err).To(BeNil())
			re := regexp.MustCompile(`ruby '(\d+.\d+.\d+)', :engine => 'jruby', :engine_version => '(\d+.\d+.\d+.\d+)'`)
			matches := re.FindStringSubmatch(string(data))
			rubyVersion = matches[1]
			jrubyVersion = matches[2]

			app = cutlass.New(dir)
			app.Memory = "512M"
		})

		It("installs the correct version of JRuby", func() {
			PushAppAndConfirm(app)

			By("the buildpack logged it installed a specific version of JRuby", func() {
				Expect(app.Stdout.String()).To(ContainSubstring("Installing openjdk"))
				Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing jruby ruby-%s-jruby-%s", rubyVersion, jrubyVersion)))
				Expect(app.GetBody("/ruby")).To(ContainSubstring(fmt.Sprintf("jruby %s", rubyVersion)))
			})

			By("the OpenJDK runs properly", func() {
				Expect(app.Stdout.String()).ToNot(ContainSubstring("OpenJDK 64-Bit Server VM warning"))
			})
		})

		Context("a cached buildpack", func() {
			BeforeEach(SkipUnlessCached)

			AssertNoInternetTraffic("sinatra_jruby")
		})
	})
	Context("with a jruby start command", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "jruby_start_command"))
			app.Memory = "512M"
		})

		It("stages and runs successfully", func() {
			PushAppAndConfirm(app)
		})
	})
})
