package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testDefault(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("when the ruby version is specified in the app", func() {
			it("builds and runs the app with the specified version", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default", "specified_ruby_version"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("ruby 3.1.6")).WithEndpoint("/ruby"))
			})
		})

		context("when the ruby version is not specified in the app", func() {
			it("builds and runs the app with the default version", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default", "unspecified_ruby_version"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(MatchRegexp(`ruby 3\.\d+\.\d+`)).WithEndpoint("/ruby"))
			})
		})

		context("rails7", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default", "rails7"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
			})
		})

		context("rails6 sprockets", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default", "rails6_sprockets"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
				Eventually(deployment).Should(Serve(MatchRegexp(`Ruby version: ruby \d+\.\d+\.\d+`)))
				Eventually(deployment).Should(Serve(MatchRegexp(`Node version: v\d+\.\d+\.\d+`)))
			})
		})

		context("vendor bundle", func() {
			it("builds and runs the app, ignoring the checked-in vendor/bundle directory", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default", "vendor_bundle"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Healthy")))
			})
		})

		context("custom gemfile", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"BUNDLE_GEMFILE": "Gemfile-APP",
					}).
					Execute(name, filepath.Join(fixtures, "default", "custom_gemfile"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World")))
			})
		})

		context("relative gemspec", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default", "relative_gemspec_path"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World")))
			})
		})

		context("jruby", func() {
			it("builds and runs the jruby sinatra app", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default", "sinatra_jruby"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(MatchRegexp(`Installing jruby \d+\.\d+\.\d+\.\d+`)))
				Eventually(deployment, 1*time.Minute, 1*time.Second).Should(Serve(ContainSubstring("jruby 3.1.4")).WithEndpoint("/ruby"), logs.String())
			})
		})
	}
}
