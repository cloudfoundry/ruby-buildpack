require "language_pack/shell_helpers"
require 'yaml'
require 'pathname'

module LanguagePack
  class RubySemverVersion
    def initialize(gemfile, manifest_file)
      ruby_requirement(gemfile)
      rubies(manifest_file)
    end

    def version
      return "" unless @ruby_requirement

      @rubies.find do |version|
        @ruby_requirement.satisfied_by? version
      end.to_s
    end

    private

    def rubies(manifest_file)
      manifest = YAML.load_file(manifest_file)
      @rubies = manifest['dependencies'].select do |hash|
        hash['name'] == 'ruby'
      end.map do |hash|
        Gem::Version.new(hash['version'])
      end.sort.reverse
    end

    def ruby_requirement(gemfile)
      # This can be restored to passing in 'gemfile' once bundler 1.13 is
      # released
      full_gemfile_path = Pathname.new(gemfile).expand_path.to_s
      ruby_version = Bundler::Dsl.evaluate(full_gemfile_path, "#{gemfile}.lock", {}).ruby_version

      if ruby_version
        engine_versions = ruby_version.engine_versions
      else
        engine_versions = "~> #{LanguagePack::RubyVersion::DEFAULT_VERSION_NUMBER}"
      end
      @ruby_requirement = Gem::Requirement.create(engine_versions)
    end
  end
end
