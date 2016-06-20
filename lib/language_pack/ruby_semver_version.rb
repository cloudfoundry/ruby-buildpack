require "language_pack/shell_helpers"
require 'yaml'

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
      ruby_version = Bundler::Dsl.evaluate(gemfile, "#{gemfile}.lock", {}).ruby_version

      if ruby_version
        engine_versions = ruby_version.engine_versions
      else
        engine_versions = "~> #{LanguagePack::RubyVersion::DEFAULT_VERSION_NUMBER}"
      end
      @ruby_requirement = Gem::Requirement.create(engine_versions)
    end
  end
end
