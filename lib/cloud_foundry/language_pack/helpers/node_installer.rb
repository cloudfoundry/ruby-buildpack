class LanguagePack::NodeInstaller
  # Get default node version from manifest
  bin_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "..", "compile-extensions", "bin"))
  manifest_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "..", "manifest.yml"))
  default_node_version = `#{bin_path}/default_version_for #{manifest_path} node`.chomp

  # Redefine default ruby version constant without warning
  const = 'MODERN_NODE_VERSION'
  self.send(:remove_const, const) if self.const_defined?(const)
  MODERN_NODE_VERSION = default_node_version

  const = 'MODERN_BINARY_PATH'
  self.send(:remove_const, const) if self.const_defined?(const)
  MODERN_BINARY_PATH = "node-v#{MODERN_NODE_VERSION}-linux-x64"
end
