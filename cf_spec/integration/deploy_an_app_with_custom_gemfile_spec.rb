require 'cf_spec_helper'

describe 'App with custom Gemfile' do
  before(:all) do
    @app = Machete.deploy_app('custom_gemfile')
  end

  after(:all) { Machete::CF::DeleteApp.new.execute(@app) }

  it 'starts' do
    expect(@app).to be_running
  end

  it 'uses the version of ruby specified in Gemfile-APP' do
    expect(@app).to have_logged 'Using Ruby version: ruby-2.2.6'
  end
end
