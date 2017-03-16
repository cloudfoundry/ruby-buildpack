require 'cf_spec_helper'

describe 'Rails 4 App' do
  subject(:app) do
    Machete.deploy_app(app_name)
  end
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'in an offline environment', :cached do
    let(:app_name) { 'rails4' }

    specify do
      expect(app).to be_running

      browser.visit_path('/')
      expect(browser).to have_body('The Kessel Run')

      expect(app).not_to have_internet_traffic
      expect(app).to have_logged /Downloaded \[file:\/\/.*\]/
    end

  end

  context 'in an online environment', :uncached do
    context 'app has dependencies' do
      let(:app_name) { 'rails4' }

      specify do
        expect(app).to be_running
        expect(app).to have_logged /Downloaded.*node-4\./

        browser.visit_path('/')
        expect(browser).to have_body('The Kessel Run')
        expect(app).to have_logged /Downloaded \[https:\/\/.*\]/
      end
    end

    context 'app has non vendored dependencies' do
      let(:app_name) { 'rails4_not_vendored' }

      specify do
        expect(Dir.exists?("cf_spec/fixtures/#{app_name}/vendor")).to eql(false)
        expect(app).to be_running

        browser.visit_path('/')
        expect(browser).to have_body('The Kessel Run')
      end

      it "uses a proxy during staging if present" do
        expect(app).to use_proxy_during_staging
      end
    end
  end
end
