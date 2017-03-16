require 'securerandom'
require "language_pack"
require "language_pack/rails5"
require "language_pack/shell_helpers"

class LanguagePack::Rails51 < LanguagePack::Rails5
  include LanguagePack::ShellHelpers

  # @return [Boolean] true if it's a Rails 5.x app
  def self.use?
    instrument "rails5.use" do
      rails_version = bundler.gem_version('railties')
      return false unless rails_version
      is_rails = rails_version >= Gem::Version.new('5.1.0.beta') &&
                 rails_version <  Gem::Version.new('6.0.0.beta1')
      return is_rails
    end
  end

  def run_assets_precompile_rake_task
    instrument "rails5.run_assets_precompile_rake_task" do
      if File.exists?('bin/yarn') && File.exists?('yarn.lock')
        run!("./bin/yarn install")
      end

      super
    end
  end
end
