# Cloud Foundry Ruby Buildpack
[![CF Slack](https://www.google.com/s2/favicons?domain=www.slack.com) Join us on Slack](https://cloudfoundry.slack.com/messages/buildpacks/)

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for Ruby based apps.

This is based on the [Heroku buildpack](https://github.com/heroku/heroku-buildpack-ruby).

Additional information can be found at [CloudFoundry.org](http://docs.cloudfoundry.org/buildpacks/).

### Buildpack User Documentation

Official buildpack documentation can be found at [ruby buildpack docs](http://docs.cloudfoundry.org/buildpacks/ruby/index.html).

### Building the Buildpack

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
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ --uncached | --cached ]
  ```

1. Use in Cloud Foundry

    Upload the buildpack to your Cloud Foundry and optionally specify it by name

    ```bash
    cf create-buildpack custom_ruby_buildpack ruby_buildpack-cached-custom.zip 1
    cf push my_app -b custom_ruby_buildpack
    ```

### Testing
Buildpacks use the [Machete](https://github.com/cloudfoundry/machete) framework for running integration tests.

To test a buildpack, run the following command from the buildpack's directory:

```
BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-build
```

More options can be found on Machete's [Github page](https://github.com/cloudfoundry/machete).

### Contributing

Find our guidelines [here](./CONTRIBUTING.md).

### Help and Support

Join the #buildpacks channel in our [Slack community](http://slack.cloudfoundry.org/) if you need any further assistance.

### Reporting Issues

Please fill out the issue template fully if you'd like to start an issue for the buildpack.

### Active Development

The project backlog is on [Pivotal Tracker](https://www.pivotaltracker.com/projects/1042066)
