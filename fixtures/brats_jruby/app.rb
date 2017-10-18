require 'sinatra'
require 'nokogiri'
require 'eventmachine'
require 'bcrypt'
require 'bson'
require 'jdbc/mysql'
require 'jdbc/postgres'

Jdbc::MySQL.load_driver
Jdbc::Postgres.load_driver

get '/' do
  'Hello, World'
end

get '/nokogiri' do
  doc = Nokogiri::XML(open('test.xml'))
  doc.xpath('//xml')
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

get '/bson' do
  1024.to_bson.unpack('H*').first
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
