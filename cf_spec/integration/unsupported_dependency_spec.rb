require 'cf_spec_helper'

describe 'CF Ruby Buildpack' do
  before(:all) do
    @app = Machete.deploy_app('unsupported_ruby')
  end

  after(:all) { Machete::CF::DeleteApp.new.execute(@app) }

  it 'should not start' do
    expect(@app).to_not be_running
  end

  it 'displays a nice error message when Ruby 99.99.99 is specified' do
    expect(@app).to have_logged 'DEPENDENCY MISSING IN MANIFEST: ruby 99.99.99'
  end
end
