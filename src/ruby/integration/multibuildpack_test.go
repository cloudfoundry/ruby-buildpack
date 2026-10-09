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

func testMultiBuildpack(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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
			if name != "" && (!settings.KeepFailedContainers || !t.Failed()) {
				Expect(platform.Delete.Execute(name)).To(Succeed())
			}
		})

		context("when ruby is a supply for the binary buildpack", func() {
			it("finds the supplied dependency in the runtime container", func() {
				deploymentProcess := platform.Deploy.
					WithBuildpacks(
						"ruby_buildpack",
						"https://github.com/cloudfoundry/binary-buildpack#master",
					)

				deployment, _, err := deploymentProcess.Execute(name, filepath.Join(fixtures, "multibuildpack", "no_gemfile"))
				Expect(err).NotTo(HaveOccurred())

				// Extended timeout for multi-buildpack deployment consistency,
				// matching the pattern established in d5294b6d.
				Eventually(deployment, 90*time.Second, 2*time.Second).Should(Serve(MatchRegexp(`Ruby Version: \d+\.\d+\.\d+`)))
			})
		})

		context("when supplied with nodejs", func() {
			it("finds the supplied dependency in the runtime container", func() {
				deployment, _, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/nodejs-buildpack#master",
						"ruby_buildpack",
					).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "rails72"))
				Expect(err).NotTo(HaveOccurred())

				// The default 20s Eventually timeout has been observed to be insufficient for
				// this heavy multi-buildpack Rails 7.2 + Webpacker app, especially under
				// parallel test execution. Staging alone takes 4-5 minutes, and app boot can
				// take 60-120s. Extended timeout matches the pattern established for JRuby
				// (#1140) and other slow-booting fixtures (d5294b6d).
				Eventually(deployment, 90*time.Second, 2*time.Second).Should(Serve(ContainSubstring("Ruby version: ruby 3.")))
				Eventually(deployment, 90*time.Second, 2*time.Second).Should(Serve(MatchRegexp(`Node version: v\d+\.\d+\.\d+`)))
			})
		})

		context("when supplied with go", func() {
			it("finds the supplied dependency in the runtime container", func() {
				deployment, _, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/go-buildpack#master",
						"ruby_buildpack",
					).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "ruby_calls_go"))
				Expect(err).NotTo(HaveOccurred())

				// Extended timeout for multi-buildpack deployment consistency.
				Eventually(deployment, 90*time.Second, 2*time.Second).Should(Serve(MatchRegexp(`RUBY_VERSION IS \d+\.\d+\.\d+`)))
				Eventually(deployment, 90*time.Second, 2*time.Second).Should(Serve(MatchRegexp(`go version go\d+\.\d+(\.\d+)?`)))
			})
		})

		context("when supplied with .NET Core", func() {
			it("finds the supplied dependency in the runtime container", func() {
				deployment, _, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/dotnet-core-buildpack#master",
						"ruby_buildpack",
					).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "dotnet_core"))
				Expect(err).NotTo(HaveOccurred())

				// Extended timeout for multi-buildpack deployment consistency.
				Eventually(deployment, 90*time.Second, 2*time.Second).Should(Serve(MatchRegexp(`dotnet: \d+\.\d+\.\d+`)))
			})
		})
	}
}
