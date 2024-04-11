class ApplicationController < ActionController::Base
  def index
    node_version = `node -v`
    ruby_version = `ruby -v`
    output = 'Hello World!' + "\n" + 'Ruby version: ' + ruby_version + 'Node version: ' + node_version
    render plain: output
  end
end
