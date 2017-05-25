class LanguagePack::DefaultVersion
  def self.for(name)
    buildpack_dir = File.expand_path(File.join(File.dirname(__FILE__), "..", ".."))
    bin_path = File.join(buildpack_dir, "compile-extensions", "bin")
    version = `#{bin_path}/default_version_for #{File.join(buildpack_dir, 'manifest.yml')} #{name}`

    unless $?.success?
      message = "\nCould not find default version for #{name}"
      raise StandardError, message
    end

    version
  end
end
