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

func testOffline(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		it("runs a vendored cached app", func() {
			deployment, _, err := platform.Deploy.
				WithoutInternetAccess().
				Execute(name, filepath.Join(fixtures, "default", "vendor_cache"))
			Expect(err).NotTo(HaveOccurred())

			// The default 20s Eventually timeout has been observed to
			// intermittently race with the Docker platform backend under
			// parallel CI load: the container reports as started before the
			// app process inside has actually finished booting and bound
			// its port, producing a raw "connection refused" at the 20s
			// mark rather than a graceful retry. Give this more headroom,
			// matching the precedent set for the JRuby fixture (#1140).
			Eventually(deployment, 90*time.Second, 2*time.Second).Should(Serve(ContainSubstring("Healthy")))
		})
	}
}
