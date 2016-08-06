require 'cf_spec_helper'

describe 'CF Ruby HTTPS Mirror' do
 
  it 'displays the error info when ENV[GEM_SOURCE_MIRROR_HTTPS] is set incorrectly' do
    app = Machete.deploy_app('mirror_https', env: {
      GEM_SOURCE_MIRROR_HTTPS: 'ruby.taobao.org'
    })
    expect(app).to_not be_running
    expect(app).to have_logged 'ArgumentError: Gem sources must be absolute.'
    Machete::CF::DeleteApp.new.execute(app)
  end
  
  it 'displays the error info when ENV[GEM_SOURCE_MIRROR_HTTPS] is unavailable' do
    app = Machete.deploy_app('mirror_https', env: {
      GEM_SOURCE_MIRROR_HTTPS: 'https://ruby.taoba0'
    })
    expect(app).to_not be_running
    expect(app).to have_logged 'Running: bundle config mirror.https://rubygems.org https://ruby.taoba0'
    expect(app).to have_logged 'Bundler::HTTPError Could not fetch specs from https://ruby.taoba0/'
    Machete::CF::DeleteApp.new.execute(app)
  end
  
  it 'replaces the source successfully when ENV[GEM_SOURCE_MIRROR_HTTPS] is set correctly' do
    app = Machete.deploy_app('mirror_https', env: {
      GEM_SOURCE_MIRROR_HTTPS: 'https://ruby.taobao.org'
    })
    expect(app).to be_running
    expect(app).to have_logged 'Running: bundle config mirror.https://rubygems.org https://ruby.taobao.org'
    expect(app).to have_logged 'Fetching gem metadata from https://ruby.taobao.org/.....'
    Machete::CF::DeleteApp.new.execute(app)
  end
  
end
