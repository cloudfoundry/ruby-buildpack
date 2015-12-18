require 'cf_spec_helper'

describe 'CF Ruby HTTP/HTTPS Mirror' do
    
  it 'replaces the sources successfully when ENV[GEM_SOURCE_MIRROR_HTTP] and ENV[GEM_SOURCE_MIRROR_HTTPS] are set correctly' do
    app = Machete.deploy_app('mirror_http_https', env: {
      GEM_SOURCE_MIRROR_HTTP: 'https://ruby.taobao.org',
      GEM_SOURCE_MIRROR_HTTPS: 'https://ruby.taobao.org'
    })
    expect(app).to be_running
    expect(app).to have_logged 'Running: bundle config mirror.http://rubygems.org https://ruby.taobao.org'
    expect(app).to have_logged 'Running: bundle config mirror.https://rubygems.org https://ruby.taobao.org'
    expect(app).to have_logged 'Fetching gem metadata from https://ruby.taobao.org/.....'
    Machete::CF::DeleteApp.new.execute(app)
  end
  
end
