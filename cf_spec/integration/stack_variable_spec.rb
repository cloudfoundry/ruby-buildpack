$: << 'cf_spec'
require 'cf_spec_helper'

describe 'Stack environment should not change' do
  let(:app_name) { 'sinatra_web_app' }

  specify do
    app = Machete.deploy_app(app_name)
    expect(app).to be_running

    Machete.push(app)
    expect(app).to be_running

    expect(app).to_not have_logged 'Changing stack from'
    expect(app).to_not have_logged 'are the same file'

    browser = Machete::Browser.new(app)
    browser.visit_path('/')
    expect(browser).to have_body('Hello world!')
  end
end
