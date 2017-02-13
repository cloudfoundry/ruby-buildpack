require 'eventmachine'
require 'em-http-server'

class HTTPHandler < EM::HttpServer::Server

  def process_http_request
    puts  @http_request_method
    puts  @http_request_uri
    puts  @http_query_string
    puts  @http_protocol
    puts  @http_content
    puts  @http[:cookie]
    puts  @http[:content_type]
    # you have all the http headers in this hash
    puts  @http.inspect

    response = EM::DelegatedHttpResponse.new(self)
    response.status = 200
    response.content_type 'text/html'
    response.content = "Hello, World!"
    response.send_response
  end

  def http_request_errback e
    # printing the whole exception
    puts e.inspect
  end

end

EM::run do
  EM::start_server("0.0.0.0", ENV['PORT'], HTTPHandler)
end
