require 'cf_spec_helper'

describe 'App with relative gemspec path' do
  before(:all) do
    @app = Machete.deploy_app('relative_gemspec_path')
  end

  after(:all) { Machete::CF::DeleteApp.new.execute(@app) }

  it 'starts' do
    expect(@app).to be_running
  end

  it 'loads the gem with the relative gemspec path' do
    expect(@app).to have_logged "Using hola 0.0.0 from source at `gems/hola`"
  end
end
