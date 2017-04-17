require 'cf_spec_helper'

describe 'JRuby App' do
  before(:all) do
    @app = Machete.deploy_app('sinatra_jruby')
    expect(@app).to be_running
    @browser = Machete::Browser.new(@app)
  end

  after(:all) do
    Machete::CF::DeleteApp.new.execute(@app)
  end

  specify 'the buildpack logged it installed a specific version of JRuby' do
    expect(@app).to have_logged 'Installing JVM'
    expect(@app).to have_logged /ruby-2.3.\d+-jruby-9.1.\d+.0/
    @browser.visit_path('/ruby')
    expect(@browser).to have_body /jruby 2.3.\d+/
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
