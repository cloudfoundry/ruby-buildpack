package brats_test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	return bratshelper.InitBpData().Marshal()
}, func(data []byte) {
	// Run on all nodes
	bratshelper.Data.Unmarshal(data)

	Expect(cutlass.CopyCfHome()).To(Succeed())
	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(bratshelper.Data.Cached, "_buildpack", "", 1))).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(bratshelper.Data.Uncached, "_buildpack", "", 1))).To(Succeed())
	Expect(os.Remove(bratshelper.Data.CachedFile)).To(Succeed())
})

func TestBrats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brats Suite")
}

func CopyBrats(rubyVersion string) *cutlass.App {
	dir, err := cutlass.CopyFixture(filepath.Join(bratshelper.Data.BpDir, "fixtures", "brats_ruby"))
	Expect(err).ToNot(HaveOccurred())
	data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
	Expect(err).ToNot(HaveOccurred())
	if rubyVersion == "" {
		manifest, err := libbuildpack.NewManifest(bratshelper.Data.BpDir, nil, time.Now())
		Expect(err).ToNot(HaveOccurred())
		dep, err := manifest.DefaultVersion("ruby")
		Expect(err).ToNot(HaveOccurred())
		rubyVersion = dep.Version
	} else if strings.Contains(rubyVersion, "x") {
		manifest, err := libbuildpack.NewManifest(bratshelper.Data.BpDir, nil, time.Now())
		Expect(err).ToNot(HaveOccurred())
		depVersions := manifest.AllDependencyVersions("ruby")
		rubyVersion, err = libbuildpack.FindMatchingVersion(rubyVersion, depVersions)
		Expect(err).ToNot(HaveOccurred())
	}
	data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
	Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())

	return cutlass.New(dir)
}

func CopyBratsJRuby(fullRubyVersion string) *cutlass.App {
	m := regexp.MustCompile(`ruby-(.*)-jruby-(.*)`).FindStringSubmatch(fullRubyVersion)
	if len(m) != 3 {
		panic("Incorrect jruby version " + fullRubyVersion)
	}
	rubyVersion := m[1]
	jrubyVersion := m[2]

	dir, err := cutlass.CopyFixture(filepath.Join(bratshelper.Data.BpDir, "fixtures", "brats_jruby"))
	Expect(err).ToNot(HaveOccurred())
	data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
	Expect(err).ToNot(HaveOccurred())
	data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
	data = bytes.Replace(data, []byte("<%= engine_version %>"), []byte(jrubyVersion), -1)
	Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())
	return cutlass.New(dir)
}

func PushApp(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(app.InstanceStates, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}
