$: << 'cf_spec'
require 'cf_spec_helper'

describe 'Rack App with Ruby 2.2.0' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:app_name) { 'ruby_2_2_0' }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'in an offline environment', if: Machete::BuildpackMode.offline? do
    specify do
      expect(app).to be_running

      expect(app).to have_logged 'Using Ruby version: ruby-2.2.0'

      browser.visit_path('/')
      expect(browser).to have_body('Hello world!')

      expect(app.host).not_to have_internet_traffic
    end
  end

  context 'in an online environment', if: Machete::BuildpackMode.online? do
    specify do
      expect(app).to be_running

      expect(app).to have_logged 'Using Ruby version: ruby-2.2.0'

      browser.visit_path('/')
      expect(browser).to have_body('Hello world!')
    end
  end
end
