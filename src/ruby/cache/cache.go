package cache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Metadata struct {
	Stack         string
	SecretKeyBase string
}

type Cache struct {
	buildDir string
	cacheDir string
	depDir   string
	names    []string
	metadata Metadata
	log      *libbuildpack.Logger
	yaml     YAML
}

type Stager interface {
	BuildDir() string
	CacheDir() string
	DepDir() string
}

type YAML interface {
	Load(file string, obj interface{}) error
	Write(dest string, obj interface{}) error
}

func New(stager Stager, log *libbuildpack.Logger, yaml YAML) (*Cache, error) {
	c := &Cache{
		buildDir: stager.BuildDir(),
		cacheDir: stager.CacheDir(),
		depDir:   filepath.Join(stager.DepDir()),
		names:    []string{"vendor_bundle", "node_modules"},
		metadata: Metadata{},
		log:      log,
		yaml:     yaml,
	}

	if err := yaml.Load(c.metadata_yml(), &c.metadata); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return c, nil
}

func (c *Cache) Metadata() *Metadata {
	return &c.metadata
}

func (c *Cache) Restore() error {
	if c.metadata.Stack == os.Getenv("CF_STACK") {
		for _, name := range c.names {
			if exists, err := libbuildpack.FileExists(filepath.Join(c.cacheDir, name)); err != nil {
				return err
			} else if exists {
				c.log.BeginStep("Restoring %s from cache", name)
				if err := os.Rename(filepath.Join(c.cacheDir, name), filepath.Join(c.depDir, name)); err != nil {
					return err
				}
			}
		}
	}
	if c.metadata.Stack != "" {
		c.log.BeginStep("Skipping restoring vendor_bundle from cache, stack changed from %s to %s", c.metadata.Stack, os.Getenv("CF_STACK"))
	}
	return os.RemoveAll(filepath.Join(c.cacheDir, "vendor_bundle"))
}

func (c *Cache) Save() error {
	for _, name := range c.names {
		if exists, err := libbuildpack.FileExists(filepath.Join(c.depDir, name)); err != nil {
			return err
		} else if exists {
			c.log.BeginStep("Saving %s to cache", name)
			cmd := exec.Command("cp", "-al", filepath.Join(c.depDir, name), filepath.Join(c.cacheDir, name))
			if output, err := cmd.CombinedOutput(); err != nil {
				c.log.Error(string(output))
				return fmt.Errorf("Could not copy %s: %v", name, err)
			}
		}
	}

	c.metadata.Stack = os.Getenv("CF_STACK")
	if err := c.yaml.Write(c.metadata_yml(), c.metadata); err != nil {
		return err
	}

	return nil
}

func (c *Cache) metadata_yml() string {
	return filepath.Join(c.cacheDir, "metadata.yml")
}
