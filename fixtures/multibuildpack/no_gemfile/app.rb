# Most bare-bones app
# webrick isn't part of stdlib in ruby 3.x
require 'socket'

server = TCPServer.new(ENV.fetch('PORT', 8080))
response = "Ruby Version: #{RUBY_VERSION}"

loop {
  client = server.accept
  headers = ["HTTP/1.1 200 OK",
            "Content-Type: text/html",
            "Content-Length: #{response.length}\r\n\r\n"].join("\r\n")
  client.puts headers
  client.puts response
  client.close
}
