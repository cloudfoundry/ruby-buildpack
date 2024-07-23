package integration_test

import (
	"path/filepath"
	"testing"

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
			Expect(platform.Delete.Execute(name)).To(Succeed())
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

				Eventually(deployment).Should(Serve(MatchRegexp(`Ruby Version: \d+\.\d+\.\d+`)))
			})
		})

		context("when supplied with nodejs", func() {
			it("finds the supplied dependency in the runtime container", func() {
				deployment, _, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/nodejs-buildpack#master",
						"ruby_buildpack",
					).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "rails6"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Ruby version: ruby 3.")))
				Eventually(deployment).Should(Serve(ContainSubstring("Node version: v18.")))
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

				Eventually(deployment).Should(Serve(MatchRegexp(`RUBY_VERSION IS \d+\.\d+\.\d+`)))
				Eventually(deployment).Should(Serve(MatchRegexp(`go version go\d+\.\d+(\.\d+)?`)))
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

				Eventually(deployment).Should(Serve(MatchRegexp(`dotnet: \d+\.\d+\.\d+`)))
			})
		})
	}
}
