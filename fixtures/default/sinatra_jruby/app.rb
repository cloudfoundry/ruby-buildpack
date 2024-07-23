require 'sinatra'

get '/' do
  'Hello world!'
end

get '/ruby' do
  "#{RUBY_ENGINE} #{RUBY_VERSION}"
end
