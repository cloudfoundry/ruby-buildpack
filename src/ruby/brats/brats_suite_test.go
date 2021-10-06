package brats_test

import (
	"bytes"
	"flag"
	"fmt"
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

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	return bratshelper.InitBpData(os.Getenv("CF_STACK"), ApiHasStackAssociation()).Marshal()
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

// jruby 9.3.X.X = ruby 2.6.X
// jruby 9.2.X.X = ruby 2.5.X
// jruby 9.1.X.X = ruby 2.3.X
func rubyVersionFromJRubyVersion(jrubyVersion string) (string, error) {
	jrubyVersionRegex := regexp.MustCompile(`^(9.\d+).\d+.\d+$`)
	version := jrubyVersionRegex.FindStringSubmatch(jrubyVersion)
	if version == nil {
		return "", fmt.Errorf("JRuby version is not of expected format: expected 9.X.X.X, got %s", jrubyVersion)
	}
	switch version[1] {
	case "9.3":
		return "~>2.6", nil
	case "9.2":
		return "~>2.5", nil
	case "9.1":
		return "~>2.3", nil
	default:
		return "", fmt.Errorf("Unknown JRuby -> Ruby version mapping for JRuby version %s", jrubyVersion)
	}
}

func CopyBratsJRuby(jrubyVersion string) *cutlass.App {
	rubyVersion, err := rubyVersionFromJRubyVersion(jrubyVersion)
	Expect(err).ToNot(HaveOccurred())

	dir, err := cutlass.CopyFixture(filepath.Join(bratshelper.Data.BpDir, "fixtures", "brats_jruby"))
	Expect(err).ToNot(HaveOccurred())

	data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
	Expect(err).ToNot(HaveOccurred())

	data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
	data = bytes.Replace(data, []byte("<%= engine_version %>"), []byte(jrubyVersion), -1)
	Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())

	Expect(createGemfileLockFile(jrubyVersion, dir)).To(Succeed())

	return cutlass.New(dir)
}

func createGemfileLockFile(jrubyVersion string, fixtureDir string) error {
	jrubyVersionRegex := regexp.MustCompile(`^(9.\d+).\d+.\d+$`)
	version := jrubyVersionRegex.FindStringSubmatch(jrubyVersion)
	if version == nil {
		return fmt.Errorf("JRuby version is not of expected format: expected 9.X.X.X, got %s", jrubyVersion)
	}

	var buffer []byte
	switch version[1] {
	case "9.2":
		buffer = []byte(`GEM
  remote: https://rubygems.org/
  specs:
    bcrypt (3.1.16-java)
    eventmachine (1.2.7-java)
    jdbc-mysql (8.0.17)
    jdbc-postgres (42.2.14)
    mustermann (1.1.1)
      ruby2_keywords (~> 0.0.1)
    nokogiri (1.11.1-java)
      racc (~> 1.4)
    racc (1.5.2-java)
    rack (2.2.3)
    rack-protection (2.1.0)
      rack
    ruby2_keywords (0.0.4)
    sinatra (2.1.0)
      mustermann (~> 1.0)
      rack (~> 2.2)
      rack-protection (= 2.1.0)
      tilt (~> 2.0)
    tilt (2.0.10)

PLATFORMS
  universal-java-11

DEPENDENCIES
  bcrypt
  eventmachine
  jdbc-mysql
  jdbc-postgres
  nokogiri
  sinatra`)
	case "9.1":
		buffer = []byte(`GEM
  remote: https://rubygems.org/
  specs:
    bcrypt (3.1.10-java)
    eventmachine (1.0.7-java)
    jdbc-mysql (5.1.35)
    jdbc-postgres (9.4.1200)
    nokogiri (1.6.6.2-java)
    rack (1.6.1)
    rack-protection (1.5.3)
      rack
    sinatra (1.4.6)
      rack (~> 1.4)
      rack-protection (~> 1.4)
      tilt (>= 1.3, < 3)
    tilt (2.0.1)

PLATFORMS
  java

DEPENDENCIES
  bcrypt
  eventmachine
  jdbc-mysql
  jdbc-postgres
  nokogiri
  sinatra`)
	default:
		return fmt.Errorf("Unknown JRuby version %s, could not write Gemfile.lock", jrubyVersion)
	}

	Expect(ioutil.WriteFile(filepath.Join(fixtureDir, "Gemfile.lock"), buffer, 0644)).To(Succeed())

	return nil
}

func PushApp(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(app.InstanceStates, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}

func ApiHasStackAssociation() bool {
	supported, err := cutlass.ApiGreaterThan("2.113.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}
