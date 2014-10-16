require 'bundler/setup'
require 'machete'
require 'machete/matchers'
require 'rspec/retry'

require 'language_pack'
require 'cloud_foundry/language_pack/fetcher'

`mkdir -p log`
Machete.logger = Machete::Logger.new("log/integration.log")
