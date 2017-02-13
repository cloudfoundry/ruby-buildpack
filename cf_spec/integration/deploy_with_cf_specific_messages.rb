require 'cf_spec_helper'

describe 'CF Specific log output' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:app_name) { 'sinatra' }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'when no Procfile is present' do
    it "does not output Procfile and default usage warning" do
      expect(app).to_not have_logged "No Procfile detected, using the default"
    end
  end
end
