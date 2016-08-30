require 'sinatra'
require 'hola'

STDOUT.sync = true

Hola.hi

get '/' do
    "Hello World!"
end

get '/version' do
    "#{RUBY_VERSION}-p#{RUBY_PATCHLEVEL}"
end

get '/sleep' do
    sleep(900)
end

get '/biglog' do
    buf = "0" * 32000
    for i in 0..1500
        puts "#{i} -> #{buf}"
    end
    'logged'
end

get '/env' do
    res = ''
    ENV.each do |k, v|
        res << "#{k}: #{v}<br/>"
    end
    res
end

get '/hello/:name' do
    "Hello! Your name is [#{params[:name]}]."
end

get '/hello' do
    "Hello! Your name is [#{params[:name]}]."
end

get '/413' do
    status 413
    body "Insert Crazy 413 message here"
end
