# Load the rails application
require File.expand_path('../application', __FILE__)

# Initialize the rails application
Rails3Mri193::Application.initialize!

Rails.logger.info 'Logging is being redirected to STDOUT with rails_log_stdout plugin'
