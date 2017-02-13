require 'sinatra'

get '/' do
  begin
    #requiring execjs RTEs if node (or a js runtime) is not available.
    require 'execjs'

    'Successfully required execjs'

  rescue => e
    logger.info e
    'Failed requiring execjs'
  end
end

