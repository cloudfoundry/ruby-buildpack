require 'cf_spec_helper'

describe 'Rails 5.1 (Webpack/Yarn) App' do
  subject(:app) do
    Machete.deploy_app(app_name)
  end
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'in an online environment' do
    let(:app_name) { 'rails51' }

    specify do
      expect(app).to be_running
      expect(app).to have_logged /Downloaded.*node-6\./

      browser.visit_path('/')
      expect(browser).to have_body('Hello World')
      expect(app).to have_logged /Started GET "\/" for/
    end
  end
end
