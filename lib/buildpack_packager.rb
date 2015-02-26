require 'base_packager'
require 'json'

class BuildpackPackager < BasePackager
  def dependencies
    [
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/node-0.4.7.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/libyaml-0.1.6.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/bundler-1.6.3.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/ruby_versions.yml',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/ruby-2.1.1.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/ruby-2.1.0.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/ruby-2.0.0.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/ruby-1.9.3.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/ruby-1.9.2.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/ruby-1.8.7.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/ruby-build-1.8.7.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/rails_log_stdout.tgz',
      'https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/rails3_serve_static_assets.tgz',
    ]
  end

  def excluded_files
    [
      /\.git/,
      /repos/,
      /spec/,
      /cf_spec/,
      /cf.Gemfile*/
    ]
  end
end

BuildpackPackager.new("ruby", ARGV.first.to_sym).package
