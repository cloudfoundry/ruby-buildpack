package integration_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/switchblade"
	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var settings struct {
	Buildpack struct {
		Version string
		Path    string
	}

	Cached       bool
	Serial       bool
	FixturesPath string
	GitHubToken  string
	Platform     string
	Stack        string
}

func init() {
	flag.BoolVar(&settings.Cached, "cached", false, "run cached buildpack tests")
	flag.BoolVar(&settings.Serial, "serial", false, "run serial buildpack tests")
	flag.StringVar(&settings.Platform, "platform", "cf", `switchblade platform to test against ("cf" or "docker")`)
	flag.StringVar(&settings.GitHubToken, "github-token", "", "use the token to make GitHub API requests")
	flag.StringVar(&settings.Stack, "stack", "cflinuxfs4", "stack to use as default when pusing apps")
}

func TestIntegration(t *testing.T) {
	var Expect = NewWithT(t).Expect

	format.MaxLength = 0
	SetDefaultEventuallyTimeout(10 * time.Second)

	root, err := filepath.Abs("./../../..")
	Expect(err).NotTo(HaveOccurred())

	fixtures := filepath.Join(root, "fixtures")

	platform, err := switchblade.NewPlatform(settings.Platform, settings.GitHubToken, settings.Stack)
	Expect(err).NotTo(HaveOccurred())

	err = platform.Initialize(
		switchblade.Buildpack{
			Name: "ruby_buildpack",
			URI:  os.Getenv("BUILDPACK_FILE"),
		},
		switchblade.Buildpack{
			Name: "override_buildpack",
			URI:  filepath.Join(fixtures, "util", "override_buildpack"),
		},
		switchblade.Buildpack{
			Name: "go_buildpack", // for the proxy test
			URI:  "https://github.com/cloudfoundry/go-buildpack/releases/download/v1.10.21/go-buildpack-cflinuxfs4-v1.10.21.zip",
		},
	)
	Expect(err).NotTo(HaveOccurred())

	proxyName, err := switchblade.RandomName()
	Expect(err).NotTo(HaveOccurred())

	proxyDeployment, _, err := platform.Deploy.
		WithBuildpacks("go_buildpack").
		Execute(proxyName, filepath.Join(fixtures, "util", "proxy"))
	Expect(err).NotTo(HaveOccurred())

	suite := spec.New("integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault(platform, fixtures))
	suite("MultiBuildpack", testMultiBuildpack(platform, fixtures))
	suite("Override", testOverride(platform, fixtures))

	if settings.Cached {
		suite("Offline", testOffline(platform, fixtures)) // working
	} else {
		suite("Cache", testCache(platform, fixtures)) // working
		suite("Proxy", testProxy(platform, fixtures, proxyDeployment.InternalURL))
	}

	suite.Run(t)

	Expect(platform.Delete.Execute(proxyName)).To(Succeed())
	Expect(os.Remove(os.Getenv("BUILDPACK_FILE"))).To(Succeed())
}
