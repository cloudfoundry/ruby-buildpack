require 'cf_spec_helper'

describe 'Rails 3 App' do
  before(:all) do
    @app = Machete.deploy_app('rails3_mri_193', with_pg: true)
    expect(@app).to be_running
  end

  after(:all) do
    Machete::CF::DeleteApp.new.execute(@app)
  end

  specify 'the app can be visited in the browser' do
    browser = Machete::Browser.new(@app)
    browser.visit_path('/')
    expect(browser).to have_body('hello')
  end

  context 'the app did not include the static asset or logging gems from Heroku' do
    specify 'the rails 3 plugins are installed automatically' do
      expect(@app).to have_file 'app/vendor/plugins/rails3_serve_static_assets/init.rb'
      expect(@app).to have_file 'app/vendor/plugins/rails_log_stdout/init.rb'
    end
  end

  context 'a cached buildpack', if: Machete::BuildpackMode.offline? do
    specify 'has not internet traffic' do 
      expect(@app.host).not_to have_internet_traffic
    end
  end

  context 'we include a rails logger message in the initializer' do
    specify 'the log message is visible in the cf cli app logging' do
      expect(@app).to have_logged 'Logging is being redirected to STDOUT with rails_log_stdout plugin'
    end
  end
end
