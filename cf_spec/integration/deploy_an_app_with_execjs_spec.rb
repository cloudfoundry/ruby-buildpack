require 'cf_spec_helper'

describe 'requiring execjs' do
  subject(:app) { Machete.deploy_app('with_execjs', env: {'BP_DEBUG' => '1'}) }

  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  specify do
    expect(app).to be_running
    expect(app).to have_logged /Downloaded.*node-4\./

    browser.visit_path('/')

    expect(app).to_not have_logged 'ExecJS::RuntimeUnavailable'
    expect(browser).to have_body('Successfully required execjs')

    browser.visit_path('/npm')
    expect(browser).to have_body(/Usage: npm <command>/)
  end
end
