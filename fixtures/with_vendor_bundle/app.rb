require 'webrick'
require 'colorized_string'

$stdout.sync = true
$stderr.sync = true

class Servlet < WEBrick::HTTPServlet::AbstractServlet
  def do_GET (request, response)
    if request.unparsed_uri == '/'
      response.status = 200
      response.body = "Healthy\nMy port: #{ENV['PORT']}\n"

      puts ColorizedString.new("This is blue").blue
      STDERR.puts ColorizedString.new("This is red").red

      puts "Normal output string"
      STDERR.puts "Normal output string"
    else
      response.status = 404
    end
  end
end

server = WEBrick::HTTPServer.new(:Port => ENV.fetch('PORT', '8080'))

server.mount '/', Servlet

trap('INT') {
  server.shutdown
}

server.start
