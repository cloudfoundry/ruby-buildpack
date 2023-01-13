require 'sinatra'
require 'eventmachine'
require 'bcrypt'
require 'jdbc/mysql'
require 'jdbc/postgres'

Jdbc::MySQL.load_driver
Jdbc::Postgres.load_driver

get '/' do
  'Hello, World'
end

get '/em' do
  body = nil
  EM.run do
    EM.next_tick do
      body = 'Hello, EventMachine'
      EM.stop
    end
  end
  body
end

get '/bcrypt' do
  BCrypt::Password.create('Hello, bcrypt')
end

get '/pg' do
  begin
    userurl = 'jdbc:postgresql://HOST/DATABASE'
    java.sql.DriverManager.get_connection(userurl, 'USERNAME', 'PASSWORD')
  rescue => e
    e.message
  end
end

get '/mysql' do
  begin
    userurl = 'jdbc:mysql://HOST/DATABASE'
    java.sql.DriverManager.get_connection(userurl, 'USERNAME', 'PASSWORD')
  rescue => e
    e.message
  end
end
