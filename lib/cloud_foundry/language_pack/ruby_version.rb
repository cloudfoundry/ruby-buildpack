module LanguagePack
  class RubyVersion
    # Get default ruby version from manifest
    bin_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "compile-extensions", "bin"))
    manifest_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "manifest.yml"))
    default_ruby_version = `#{bin_path}/default_version_for #{manifest_path} ruby`.chomp

    # Redefine default ruby version and default version constants without warning
    const = 'DEFAULT_VERSION_NUMBER'
    self.send(:remove_const, const) if self.const_defined?(const)
    DEFAULT_VERSION_NUMBER = default_ruby_version

    const = 'DEFAULT_VERSION'
    self.send(:remove_const, const) if self.const_defined?(const)
    DEFAULT_VERSION = "ruby-#{DEFAULT_VERSION_NUMBER}"
  end
end
