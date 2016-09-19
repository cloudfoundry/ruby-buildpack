require 'sinatra'
require 'yaml'

get '/' do
  'Hello world!'
end

get '/yaml' do
  YAML.load("{foo: [bar, baz, quux]}").to_yaml
end
