puts "-------> Buildpack version #{`cat #{File.dirname(__FILE__)}/../../../VERSION`}"

DEPENDENCIES_PATH = File.expand_path("../../dependencies", File.expand_path($0))

require 'cloud_foundry/language_pack/fetcher'
require 'cloud_foundry/language_pack/ruby'
require 'cloud_foundry/language_pack/helpers/plugins_installer'
require 'cloud_foundry/language_pack/helpers/readline_symlink'
require 'cloud_foundry/language_pack/helpers/filename_translator'

ENV['STACK'] ||= ''
