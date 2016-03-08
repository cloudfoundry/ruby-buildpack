require 'cf_spec_helper'

describe 'Rails 3 App' do
  before(:all) do
    @app = Machete.deploy_app('rails3_mri_200', env: {
      DATABASE_URL: 'sqlite3://db/test.db'
    })
    expect(@app).to be_running
    @browser = Machete::Browser.new(@app)
  end

  after(:all) do
    Machete::CF::DeleteApp.new.execute(@app)
  end

  specify 'the app can be visited in the browser' do
    @browser.visit_path('/')
    expect(@browser).to have_body('hello')
  end

  context 'the app did not include the static asset or logging gems from Heroku' do
    specify 'the rails 3 plugins are installed automatically' do
      expect(@app).to have_file '/app/vendor/plugins/rails3_serve_static_assets/init.rb'
      expect(@app).to have_file '/app/vendor/plugins/rails_log_stdout/init.rb'
    end
  end

  context 'a cached buildpack', :cached do
    specify 'has no internet traffic' do
      expect(@app).not_to have_internet_traffic
    end
  end

  context 'we include a rails logger message in the initializer' do
    specify 'the log message is visible in the cf cli app logging' do
      expect(@app).to have_logged 'Logging is being redirected to STDOUT with rails_log_stdout plugin'
    end
  end

  context 'we include a static asset' do
    specify 'app serves the static asset' do
      expect do
        @browser.visit_path('/assets/application.css')
      end.not_to raise_error
      expect(@browser).to have_body('body{color:red}')
    end
  end
end
