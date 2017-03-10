require 'cf_spec_helper'

describe 'CF Ruby Buildpack' do
  before(:all) do
    @app = Machete.deploy_app('unspecified_ruby', env: {'BP_DEBUG' => '1'})
  end

  after(:all) { Machete::CF::DeleteApp.new.execute(@app) }

  it 'starts' do
    expect(@app).to be_running
  end

  it 'uses the default ruby version when ruby version is not specified' do
    bin_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "compile-extensions", "bin"))
    manifest_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "manifest.yml"))
    default_ruby_version = `#{bin_path}/default_version_for #{manifest_path} ruby`.chomp

    expect(@app).to have_logged("Using Ruby version: ruby-#{default_ruby_version}")
  end

  it 'pulls the default version from the manifest for ruby' do
    expect(@app).to have_logged('DEBUG: default_version_for ruby is')
  end
end
