require "language_pack/shell_helpers"
require "language_pack/ruby_semver_version"
require 'yaml'

module LanguagePack
  class RubyVersion
    class BadVersionError < BuildpackError
      def initialize(output = "")
        msg = ""
        msg << output
        msg << "Can not parse Ruby Version:\n"
        msg << "Valid versions listed on: http://docs.cloudfoundry.org/buildpacks/ruby/ruby-tips.html\n"
        super msg
      end
    end

    DEFAULT_VERSION_NUMBER = "TO_BE_REPLACED_BY_CF_DEFAULTS"
    DEFAULT_VERSION        = "ruby-#{DEFAULT_VERSION_NUMBER}"
    RUBY_VERSION_REGEX     = %r{
        (?<ruby_version>\d+\.\d+\.\d+){0}
        (?<patchlevel>p-?\d+){0}
        (?<engine>\w+){0}
        (?<engine_version>.+){0}

        ruby-\g<ruby_version>(-\g<patchlevel>)?(-\g<engine>-\g<engine_version>)?
      }x

    attr_reader :set, :version, :version_without_patchlevel, :patchlevel, :engine, :ruby_version, :engine_version
    include LanguagePack::ShellHelpers

    def initialize(bundler_output, app = {})
      @set            = nil
      @bundler_output = bundler_output
      @app            = app
      set_version
      parse_version

      update_version if engine == :ruby

      @version_without_patchlevel = @version.sub(/-p-?\d+/, '')
    end

    # https://github.com/bundler/bundler/issues/4621
    def version_for_download
      version_without_patchlevel
    end

    def default?
      @version == none
    end

    # determine if we're using jruby
    # @return [Boolean] true if we are and false if we aren't
    def jruby?
      engine == :jruby
    end

    # determine if we're using rbx
    # @return [Boolean] true if we are and false if we aren't
    def rbx?
      engine == :rbx
    end

    # convert to a Gemfile ruby DSL incantation
    # @return [String] the string representation of the Gemfile ruby DSL
    def to_gemfile
      if @engine == :ruby
        "ruby '~> #{ruby_version}'"
      else
        "ruby '~> #{ruby_version}', :engine => '#{engine}', :engine_version => '#{engine_version}'"
      end
    end

    private

    def none
      if @app[:last_version] && !@app[:is_new]
        @app[:last_version]
      else
        DEFAULT_VERSION
      end
    end

    def set_version
      if @bundler_output.empty?
        @set     = false
        @version = none
      else
        @set     = :gemfile
        @version = @bundler_output
      end
    end

    def parse_version
      md = RUBY_VERSION_REGEX.match(version)
      raise BadVersionError.new("'#{version}' is not valid") unless md
      @ruby_version   = md[:ruby_version]
      @patchlevel     = md[:patchlevel]
      @engine_version = md[:engine_version] || @ruby_version
      @engine         = (md[:engine]        || :ruby).to_sym
    end

    def update_version
      manifest = File.join(File.dirname(__FILE__), '..', '..', 'manifest.yml')
      gemfile = ENV['BUNDLE_GEMFILE'] || "./Gemfile"
      version = LanguagePack::RubySemverVersion.new(gemfile,manifest).version
      return if version.empty?

      @ruby_version = version
      @version = "ruby-"+@ruby_version
    end

    def self.default_ruby_version
      bin_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "compile-extensions", "bin"))
      manifest_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "manifest.yml"))
      `#{bin_path}/default_version_for #{manifest_path} ruby`.chomp
    end
  end
end
