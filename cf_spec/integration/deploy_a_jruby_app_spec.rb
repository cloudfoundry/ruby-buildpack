require 'cf_spec_helper'

describe 'JRuby App' do
  before(:all) do
    @app = Machete.deploy_app('sinatra_jruby_web_app')
    expect(@app).to be_running
  end

  after(:all) do
    Machete::CF::DeleteApp.new.execute(@app)
  end

  specify 'the buildpack logged it installed a specific version of JRuby' do
    expect(@app).to have_logged 'Installing JVM'
    expect(@app).to have_logged 'ruby-2.0.0-jruby-1.7.24'
  end

  specify 'the OpenJDK runs properly' do
    expect(@app).not_to have_logged 'OpenJDK 64-Bit Server VM warning'
  end

  context 'a cached buildpack', :cached do
    specify 'has no internet traffic' do
      expect(@app).not_to have_internet_traffic
    end
  end
end
