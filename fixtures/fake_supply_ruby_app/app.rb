require 'sinatra'
require 'yaml'

get '/' do
  "dotnet: " + `dotnet --version`
end
