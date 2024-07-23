require 'sinatra'
require 'open3'

get '/' do
  stdout, _ = Open3.capture2e('go version')
  "RUBY_VERSION IS #{RUBY_VERSION}\n ruby -v is #{`ruby -v`}\n" + stdout
end