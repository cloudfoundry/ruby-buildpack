require 'cf_spec_helper'

describe 'app using system yaml library' do
  before(:all) do
    @app = Machete.deploy_app('sinatra')
    @browser =  Machete::Browser.new(@app)
  end

  after(:all) { Machete::CF::DeleteApp.new.execute(@app) }

  it 'starts' do
    expect(@app).to be_running
  end

  it 'displays metasyntactic variables as yaml' do
    @browser.visit_path '/yaml'
    expect(@browser).to have_body(<<~HEREDOC)
                                     ---
                                     foo:
                                     - bar
                                     - baz
                                     - quux
                                     HEREDOC
  end
end
