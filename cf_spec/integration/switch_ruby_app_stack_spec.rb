$: << 'cf_spec'
require 'cf_spec_helper'

describe 'the app is restaged with a different rootfs' do
  subject(:app) { Machete.deploy_app(app_name, stack: 'lucid64') }
  let(:app_name) { 'ruby_with_stack_compiled_dependencies' }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  it 'does not use the app cache' do
    expect(app).to be_running(60)

    browser.visit_path('/')
    expect(browser).to have_body('Hello, World!')

    replacement_app = Machete::App.new(app_name, Machete::Host.create, stack: 'cflinuxfs2')

    app_push_command = Machete::CF::PushApp.new
    app_push_command.execute(replacement_app)

    expect(replacement_app).to be_running(60)

    browser.visit_path('/')
    expect(browser).to have_body('Hello, World!')
    expect(app).to have_logged('Fetching gem metadata from https://rubygems.org/')

  end
end
