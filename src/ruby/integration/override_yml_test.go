package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("override yml", func() {
	var app *cutlass.App
	var buildpackName string
	AfterEach(func() {
		if buildpackName != "" {
			cutlass.DeleteBuildpack(buildpackName)
		}
		app = DestroyApp(app)
	})

	BeforeEach(func() {
		if !ApiHasMultiBuildpack() {
			Skip("Multi buildpack support is required")
		}

		buildpackName = "override_yml_" + cutlass.RandStringRunes(5)
		Expect(cutlass.CreateOrUpdateBuildpack(buildpackName, Fixtures("overrideyml_bp"), "")).To(Succeed())

		app = cutlass.New(Fixtures("with_execjs"))
		app.Buildpacks = []string{buildpackName + "_buildpack", "ruby_buildpack"}
	})

	It("Forces nodejs from override buildpack, installs ruby from ruby buildpack", func() {
		Expect(app.Push()).ToNot(Succeed())
		Expect(app.Stdout.String()).To(ContainSubstring("-----> OverrideYML Buildpack"))
		Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

		Eventually(app.Stdout.String).Should(ContainSubstring("-----> Installing ruby"))
		Eventually(app.Stdout.String).Should(ContainSubstring("-----> Installing node 99.99.99"))

		Eventually(app.Stdout.String).Should(MatchRegexp("Copy .*/node.tgz"))
		Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install node: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc"))
	})
})
