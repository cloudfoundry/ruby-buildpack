require "language_pack"
require "language_pack/rails2"

# Rails 3 Language Pack. This is for all Rails 3.x apps.
class LanguagePack::Rails3 < LanguagePack::Rails2
  # detects if this is a Rails 3.x app
  # @return [Boolean] true if it's a Rails 3.x app
  def self.use?
    instrument "rails3.use" do
      rails_version = bundler.gem_version('railties')
      return false unless rails_version
      is_rails3 = rails_version >= Gem::Version.new('3.0.0') &&
                  rails_version <  Gem::Version.new('4.0.0')
      return is_rails3
    end
  end

  def name
    "Ruby/Rails"
  end

  def default_process_types
    instrument "rails3.default_process_types" do
      # let's special case thin here
      web_process = bundler.has_gem?("thin") ?
        "bundle exec thin start -R config.ru -e $RAILS_ENV -p $PORT" :
        "bundle exec rails server -p $PORT"

      super.merge({
        "web" => web_process,
        "console" => "bundle exec rails console"
      })
    end
  end

  def rake_env
    default_env_vars.merge("RAILS_GROUPS" => "assets").merge(super)
  end

private

  def install_plugins
    instrument "rails3.install_plugins" do
      return false if bundler.has_gem?('rails_12factor')
      plugins = {"rails_log_stdout" => "rails_stdout_logging", "rails3_serve_static_assets" => "rails_serve_static_assets" }.
                 reject { |plugin, gem| bundler.has_gem?(gem) }
      return false if plugins.empty?
      plugins.each do |plugin, gem|
        warn "Injecting plugin '#{plugin}'"
      end
      warn "Add 'rails_12factor' gem to your Gemfile to skip plugin injection"
      install_plugin_serve_static_assets if plugins['rails3_serve_static_assets']
      install_plugin_stdout_logger if plugins['rails_log_stdout']
    end
  end

  def install_plugin_serve_static_assets
    directory = Pathname.new("vendor/plugins/rails3_serve_static_assets")
    return true if directory.exist?
    directory.mkpath
    File.write(directory.join("init.rb"), "Rails.application.class.config.serve_static_assets = true\n")
  end

  def install_plugin_stdout_logger
    directory = Pathname.new("vendor/plugins/rails_log_stdout")
    return true if directory.exist?
    directory.mkpath
    File.write(directory.join("init.rb"), %q{
begin
  STDOUT.sync = true

  def Rails.cloudfoundry_stdout_logger
    logger = Logger.new(STDOUT)
    logger = ActiveSupport::TaggedLogging.new(logger) if defined?(ActiveSupport::TaggedLogging)
    level = ENV['LOG_LEVEL'].to_s.upcase
    level = 'INFO' unless %w[DEBUG INFO WARN ERROR FATAL UNKNOWN].include?(level)
    logger.level = Logger.const_get(level)
    logger
  end

  Rails.logger = Rails.application.config.logger = Rails.cloudfoundry_stdout_logger
rescue Exception => ex
  puts %Q{WARNING: Exception during rails_log_stdout init: #{ex.message}}
end
    })
  end

  # runs the tasks for the Rails 3.1 asset pipeline
  def run_assets_precompile_rake_task
    instrument "rails3.run_assets_precompile_rake_task" do
      log("assets_precompile") do
        if File.exists?("public/assets/manifest.yml")
          puts "Detected manifest.yml, assuming assets were compiled locally"
          return true
        end

        precompile = rake.task("assets:precompile")
        return true unless precompile.is_defined?

        topic("Preparing app for Rails asset pipeline")

        precompile.invoke(env: rake_env)

        if precompile.success?
          log "assets_precompile", :status => "success"
          puts "Asset precompilation completed (#{"%.2f" % precompile.time}s)"
        else
          precompile_fail(precompile.output)
        end
      end
    end
  end

  def rake_env
    if user_env_hash.empty?
      default_env = {
        "RAILS_GROUPS" => ENV["RAILS_GROUPS"] || "assets",
        "RAILS_ENV"    => ENV["RAILS_ENV"]    || "production",
        "DATABASE_URL" => database_url
      }
    else
      default_env = {
        "RAILS_GROUPS" => "assets",
        "RAILS_ENV"    => "production",
        "DATABASE_URL" => database_url
      }
    end
    default_env.merge(user_env_hash)
  end

  # generate a dummy database_url
  def database_url
    instrument "rails3.setup_database_url_env" do
      # need to use a dummy DATABASE_URL here, so rails can load the environment
      return env("DATABASE_URL") if env("DATABASE_URL")
      scheme =
        if bundler.has_gem?("pg") || bundler.has_gem?("jdbc-postgres")
          "postgres"
      elsif bundler.has_gem?("mysql")
        "mysql"
      elsif bundler.has_gem?("mysql2")
        "mysql2"
      elsif bundler.has_gem?("sqlite3") || bundler.has_gem?("sqlite3-ruby")
        "sqlite3"
      end
      "#{scheme}://user:pass@127.0.0.1/dbname"
    end
  end
end
