# Cloud Foundry buildpack: Ruby

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for Ruby based apps.

This is based on the [Heroku buildpack] (https://github.com/heroku/heroku-buildpack-ruby).

Additional information can be found at [CloudFoundry.org](http://docs.cloudfoundry.org/buildpacks/).

## Usage

This buildpack will be used if your app has a `Gemfile` and `Gemfile.lock` in the root directory. It will then use Bundler to install your dependencies.

```bash
cf push my_app -b https://github.com/cloudfoundry/buildpack-ruby.git
```

## Disconnected environments
To use this buildpack on Cloud Foundry, where the Cloud Foundry instance limits some or all internet activity, please read the [Disconnected Environments documentation](https://github.com/cf-buildpacks/buildpack-packager/blob/master/doc/disconnected_environments.md).

### Vendoring app dependencies
As stated in the [Disconnected Environments documentation](https://github.com/cf-buildpacks/buildpack-packager/blob/master/doc/disconnected_environments.md), your application must 'vendor' it's dependencies.

For the Ruby buildpack, use bundler:

```shell 
cd <your app dir>
bundle package --all
```

```cf push``` uploads your vendored dependencies. The buildpack will compile any dependencies requiring compilation while staging your application.

## Building

The buildpack only supports the two most stable patches for each dependency in the [manifest.yml](manifest.yml).

1. Make sure you have fetched submodules

  ```bash
  git submodule update --init
  ```

1. Get latest buildpack dependencies

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle
  ```

1. Build the buildpack

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ uncached | cached ]
  ```

1. Use in Cloud Foundry

    Upload the buildpack to your Cloud Foundry and optionally specify it by name
        
    ```bash
    cf create-buildpack custom_ruby_buildpack ruby_buildpack-cached-custom.zip 1
    cf push my_app -b custom_ruby_buildpack
    ```  

### Deprecated Versions

If you would like to build the buildpack with previously supported dependency versions, you can update `manifest.yml` with entries from `.deprecated.manifest.yml`, and re-build.

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## Reporting Issues

Open an issue on this project

## Active Development

The project backlog is on [Pivotal Tracker](https://www.pivotaltracker.com/projects/1042066)
