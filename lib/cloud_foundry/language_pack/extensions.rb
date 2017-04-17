puts "-------> Buildpack version #{`cat #{File.dirname(__FILE__)}/../../../VERSION`}"

DEPENDENCIES_PATH = File.expand_path("../../dependencies", File.expand_path($0))

require 'cloud_foundry/language_pack/ruby_version'

ENV['STACK'] ||= ''
