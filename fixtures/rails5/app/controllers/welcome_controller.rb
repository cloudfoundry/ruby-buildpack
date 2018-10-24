class WelcomeController < ApplicationController
  def index
    @node_version = `node -v`
    @ruby_version = `ruby -v`
    @node_dirs = `ls -ld $DEPS_DIR/*/node*`.to_s.split(/[\r\n]+/)
  end
end

