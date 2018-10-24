require 'sinatra'
require 'yaml'

get '/' do
  supplied_file = Dir["#{ENV['HOME']}/../deps/*/supplied"].first
  supplied_file2 = Dir["#{ENV['HOME']}/../deps/*/supplied2"].first

  output = ""

  output += File.read(supplied_file).strip if supplied_file
  output += File.read(supplied_file2).strip if supplied_file2

  output
end
