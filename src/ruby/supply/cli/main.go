package main

import (

	// _ "ruby/hooks"
	"os"
	"path/filepath"
	"ruby/cache"
	"ruby/supply"
	"ruby/versions"
	"time"

	"github.com/cloudfoundry/libbuildpack"
)

func main() {
	logger := libbuildpack.NewLogger(os.Stdout)

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		logger.Error("Unable to determine buildpack directory: %s", err.Error())
		os.Exit(9)
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		logger.Error("Unable to load buildpack manifest: %s", err.Error())
		os.Exit(10)
	}

	stager := libbuildpack.NewStager(os.Args[1:], logger, manifest)
	if err := stager.CheckBuildpackValid(); err != nil {
		os.Exit(11)
	}

	err = libbuildpack.RunBeforeCompile(stager)
	if err != nil {
		logger.Error("Before Compile: %s", err.Error())
		os.Exit(12)
	}

	if err := os.MkdirAll(filepath.Join(stager.DepDir(), "bin"), 0755); err != nil {
		logger.Error("Unable to create bin directory: %s", err.Error())
		os.Exit(13)
	}

	err = stager.SetStagingEnvironment()
	if err != nil {
		logger.Error("Unable to setup environment variables: %s", err.Error())
		os.Exit(14)
	}

	cacher, err := cache.New(stager, logger, libbuildpack.NewYAML())
	if err != nil {
		logger.Error("Unable to create cacher: %s", err.Error())
		os.Exit(14)
	}

	s := supply.Supplier{
		Stager:   stager,
		Manifest: manifest,
		Log:      logger,
		Versions: versions.New(stager.BuildDir(), manifest),
		Cache:    cacher,
		Command:  &libbuildpack.Command{},
	}

	err = supply.Run(&s)
	if err != nil {
		os.Exit(15)
	}

	if err := stager.WriteConfigYml(nil); err != nil {
		logger.Error("Error writing config.yml: %s", err.Error())
		os.Exit(16)
	}
}
