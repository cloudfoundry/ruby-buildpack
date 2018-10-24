package integration_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("running supply buildpacks before the ruby buildpack", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("the app is pushed once", func() {
		BeforeEach(func() {
			if ok, err := cutlass.ApiGreaterThan("2.65.1"); err != nil || !ok {
				Skip("API version does not have multi-buildpack support")
			}

			app = cutlass.New(filepath.Join(bpDir, "fixtures", "fake_supply_ruby_app"))
			app.Buildpacks = []string{
				"https://github.com/cloudfoundry/dotnet-core-buildpack#develop",
				"ruby_buildpack",
			}
			app.Disk = "1G"
		})

		It("finds the supplied dependency in the runtime container", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("Supplying Dotnet Core"))
			Expect(app.GetBody("/")).To(MatchRegexp(`dotnet: \d+\.\d+\.\d+`))
		})
	})

	Context("an app is pushed multiple times", func() {
		var tmpDir, randomRunes string

		BeforeEach(func() {
			if ok, err := cutlass.ApiGreaterThan("2.65.1"); err != nil || !ok {
				Skip("API version does not have multi-buildpack support")
			}

			var err error
			tmpDir, err = cutlass.CopyFixture(filepath.Join(bpDir, "fixtures", "test_cache_ruby_app"))
			Expect(err).To(BeNil())
			app = cutlass.New(tmpDir)

			randomRunes = cutlass.RandStringRunes(32)
			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "RANDOM_NUMBER"), []byte(randomRunes), 0644)).To(Succeed())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		It("pushes successfully both times with same buildpacks", func() {
			app.Buildpacks = []string{
				"https://buildpacks.cloudfoundry.org/fixtures/supply-cache-new.zip",
				"ruby_buildpack",
			}
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring(randomRunes))

			Expect(ioutil.WriteFile(filepath.Join(tmpDir, "RANDOM_NUMBER"), []byte("some string"), 0644)).To(Succeed())
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring(randomRunes))
		})

		It("pushes successfully both times with diffenent non-final buildpacks", func() {
			app.Buildpacks = []string{
				"https://buildpacks.cloudfoundry.org/fixtures/supply-cache-new.zip",
				"https://buildpacks.cloudfoundry.org/fixtures/num-cache-new.zip",
				"ruby_buildpack",
			}
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring(randomRunes))
			Expect(app.Stdout.String()).To(ContainSubstring("THERE ARE 3 CACHE DIRS"))

			app.Buildpacks = []string{
				"https://buildpacks.cloudfoundry.org/fixtures/num-cache-new.zip",
				"ruby_buildpack",
			}
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("supply2"))
			Expect(app.Stdout.String()).To(ContainSubstring("THERE ARE 2 CACHE DIRS"))
		})
	})
})
