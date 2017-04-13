require 'cf_spec_helper'

describe 'App with windows Gemfile' do
  before(:all) do
    @app = Machete.deploy_app('windows')
  end

  after(:all) { Machete::CF::DeleteApp.new.execute(@app) }

  it 'warned the user about Windows line endings for windows Gemfile' do
    expect(@app).to be_running
    expect(@app).to have_logged 'WARNING: Windows line endings detected in Gemfile. Your app may fail to stage. Please use UNIX line endings.'
  end
end
