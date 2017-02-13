require 'cf_spec_helper'

describe 'Version output' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:app_name) { 'sinatra' }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'in an online environment' do
    specify do
      expect(app).to have_logged "-------> Buildpack version "
    end
  end
end
