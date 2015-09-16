require 'cf_spec_helper'

describe 'Rails 4 App' do
  subject(:app) do
    Machete.deploy_app(app_name, env: {
      DATABASE_URL: 'sqlite3://db/test.db'
    })
  end
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'in an offline environment', :cached do
    let(:app_name) { 'rails4_web_app' }

    specify do
      expect(app).to be_running

      browser.visit_path('/')
      expect(browser).to have_body('The Kessel Run')

      expect(app.host).not_to have_internet_traffic
      expect(app).to have_logged /Downloaded \[file:\/\/.*\]/
    end

  end

  context 'in an online environment', :uncached do
    context 'app has dependencies' do
      let(:app_name) { 'rails4_web_app' }

      specify do
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('The Kessel Run')
        expect(app).to have_logged /Downloaded \[https:\/\/.*\]/
      end
    end

    context 'app has no dependencies' do
      let(:app_name) { 'rails4_web_app_without_vendored_dependencies' }

      specify do
        expect(Dir.exists?("cf_spec/fixtures/#{app_name}/vendor")).to eql(false)
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('The Kessel Run')
      end
    end
  end
end
