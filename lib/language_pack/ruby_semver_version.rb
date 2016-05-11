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
      gemfile_reader = GemfileReader.new
      gemfile_reader.instance_eval(File.read(gemfile), gemfile)
      ruby_version = gemfile_reader.ruby_version
      return unless ruby_version
      @ruby_requirement = Gem::Requirement.create(ruby_version)
    end

    class GemfileReader < BasicObject
      attr_reader :ruby_version

      def ruby(*ruby_version)
        ruby_version.pop if ruby_version.last.is_a?(::Hash)
        @ruby_version = ruby_version.flatten
      end

      def method_missing *args
      end
    end
  end
end
