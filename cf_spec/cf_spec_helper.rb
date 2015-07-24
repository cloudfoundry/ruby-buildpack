require 'bundler/setup'
require 'machete'
require 'machete/matchers'

require 'language_pack'
require 'cloud_foundry/language_pack/fetcher'

`mkdir -p log`
Machete.logger = Machete::Logger.new("log/integration.log")

RSpec.configure do |config|
  config.filter_run_excluding :cached => ENV['BUILDPACK_MODE'] == 'uncached'
  config.filter_run_excluding :uncached => ENV['BUILDPACK_MODE'] == 'cached'
end
