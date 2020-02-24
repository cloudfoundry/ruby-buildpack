package finalize

import (
	"github.com/blang/semver"
)

type Versions interface {
	HasGem(string) (bool, error)
	GemMajorVersion(string) (int, error)
	HasGemVersion(string, ...string) (bool, error)
}

func (f *Finalizer) GenerateReleaseYaml() (map[string]map[string]string, error) {
	hasThin, err := f.Versions.HasGem("thin")
	if err != nil {
		return nil, err
	}
	hasRails4, err := f.Versions.HasGemVersion("rails", ">=4.0.0-beta")
	if err != nil {
		return nil, err
	}
	hasRails3, err := f.Versions.HasGemVersion("rails", ">=3.0.0")
	if err != nil {
		return nil, err
	}
	hasRails2, err := f.Versions.HasGemVersion("rails", ">=2.0.0")
	if err != nil {
		return nil, err
	}
	hasRack, err := f.Versions.HasGem("rack")
	if err != nil {
		return nil, err
	}
	processTypes := map[string]string{}
	switch {
	case hasRails4:
		processTypes["web"] = "bin/rails server -b 0.0.0.0 -p $PORT -e $RAILS_ENV"
	case hasRails3 && hasThin:
		processTypes["web"] = "bundle exec thin start -R config.ru -e $RAILS_ENV -p $PORT"
	case hasRails3:
		processTypes["web"] = "bundle exec rails server -p $PORT"
	case hasRails2 && hasThin:
		processTypes["web"] = "bundle exec thin start -e $RAILS_ENV -p $PORT"
	case hasRails2:
		processTypes["web"] = "bundle exec ruby script/server -p $PORT"
	case hasRack && hasThin:
		processTypes["web"] = "bundle exec thin start -R config.ru -e $RACK_ENV -p $PORT"
	case hasRack:
		processTypes["web"] = "bundle exec rackup config.ru -p $PORT"
	}

	return map[string]map[string]string{
		"default_process_types": processTypes,
	}, nil
}

func mustParse(s string) semver.Version {
	semver, err := semver.ParseTolerant(s)
	if err != nil {
		panic(err)
	}
	return semver
}
