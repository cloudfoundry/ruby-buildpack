require 'webrick'

server = WEBrick::HTTPServer.new(Port: ENV.fetch('PORT', 8080))
server.mount_proc '/' do |req, res|
  res.body = "Ruby Version: #{RUBY_VERSION}"
end

trap('INT') { server.stop }
server.start
