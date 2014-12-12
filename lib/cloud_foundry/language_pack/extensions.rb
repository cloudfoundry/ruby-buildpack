puts "-------> Buildpack version #{`cat #{File.dirname(__FILE__)}/../../../VERSION`}"

DEPENDENCIES_PATH = File.expand_path("../../dependencies", File.expand_path($0))

require 'cloud_foundry/language_pack/ruby'
require 'cloud_foundry/language_pack/fetcher'
require 'cloud_foundry/language_pack/helpers/readline_symlink'

ENV['STACK'] ||= ''
