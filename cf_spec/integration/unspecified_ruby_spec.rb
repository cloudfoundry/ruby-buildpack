require 'cf_spec_helper'

describe 'CF Ruby Buildpack' do
  before(:all) do
    @app = Machete.deploy_app('unspecified_ruby')
  end

  after(:all) { Machete::CF::DeleteApp.new.execute(@app) }

  it 'starts' do
    expect(@app).to be_running
  end

  it 'uses the default ruby version' do
    expect(@app).to have_logged 'Using Ruby version: ruby-2.3.0'
  end
end
