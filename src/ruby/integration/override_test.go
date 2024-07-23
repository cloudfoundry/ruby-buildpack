package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testOverride(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect = NewWithT(t).Expect

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

		it("forces node from override buildpack", func() {
			_, logs, err := platform.Deploy.
				WithBuildpacks("override_buildpack", "ruby_buildpack").
				Execute(name, filepath.Join(fixtures, "default", "rails6_sprockets"))
			Expect(err).To(MatchError(ContainSubstring("App staging failed")))

			Expect(logs).To(ContainLines(ContainSubstring("-----> OverrideYML Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("-----> Installing node")))
			Expect(logs).To(ContainLines(MatchRegexp("Copy .*/node.tgz")))
			Expect(logs).To(ContainLines(MatchRegexp(`Unable to install node: dependency sha256 mismatch: expected sha256 062d906.*, actual sha256 b56b58a.*`)))
		})
	}
}
