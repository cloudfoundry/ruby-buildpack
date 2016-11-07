require 'cf_spec_helper'

describe 'requiring execjs' do
  subject(:app) { Machete.deploy_app('app_with_execjs') }

  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  specify do
    expect(app).to be_running

    browser.visit_path('/')

    expect(app).to_not have_logged 'ExecJS::RuntimeUnavailable'
    expect(browser).to have_body('Successfully required execjs')
  end
end
