# Cloud Foundry buildpack: Ruby

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for Ruby based apps.

This is based on the [Heroku buildpack] (https://github.com/heroku/heroku-buildpack-ruby).

Additional information can be found at [CloudFoundry.org](http://docs.cloudfoundry.org/buildpacks/).

## Usage

This buildpack will be used if your app has a `Gemfile` and `Gemfile.lock` in the root directory. It will then use Bundler to install your dependencies.

```bash
cf push my_app -b https://github.com/cloudfoundry/buildpack-ruby.git
```

## Cloud Foundry Extensions - Cached Dependencies

The primary purpose of extending the heroku buildpack is to cache system dependencies for partially or fully
disconnected environments.
Historically, this was called 'offline' mode.
It is now called 'Cached dependencies'.

Cached buildpacks can be used in any environment where you would prefer the dependencies to be cached instead of fetched
from the internet.

The list of what is cached is maintained in [the manifest](manifest.yml). For a description of the manifest file,
see the [buildpack packager documentation](https://github.com/cf-buildpacks/buildpack-packager/blob/master/README.md#manifest)

The buildpack consumes cached system dependencies during staging by translating remote urls. 
In this buildpack this is specifically achieved by monkey-patching 
[the fetcher module](lib/cloud_foundry/language_pack/fetcher.rb#L14).

### App Dependencies in Cached Mode
Cached (offline) mode expects each app to [vendor its dependencies using Bundler](http://bundler.io/v1.1/bundle_package.html). The alternative is to [set up a local rubygems server](http://guides.rubygems.org/run-your-own-gem-server).

## Building

1. Make sure you have fetched submodules

  ```shell
  git submodule update --init
  ```

1. Get latest buildpack dependencies
  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle
  ```

1. Build the buildpack

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ online | offline ]
  ```

1. Use in Cloud Foundry

    Either:

    Fork the repository, push your changes and specify the path when deploying your app
    ```shell
    cf push my_app --buildpack <new buildpack repository>
    ```

    OR

    Upload the buildpack to your Cloud Foundry and specify it by name

    ```shell
    cf create-buildpack custom_ruby_buildpack ruby_buildpack-offline-custom.zip 1
    ```

## Contributing

### Run the Tests

See the [Machete](https://github.com/cf-buildpacks/machete) CF buildpack test framework for more information.

### Pull Requests

1. Fork the project
1. Submit a pull request

## Reporting Issues

Open an issue on this project

## Active Development

The project backlog is on [Pivotal Tracker](https://www.pivotaltracker.com/projects/1042066)
