class LanguagePack::NodeInstaller
  def initialize(stack)
  end

  def install
    Dir.chdir("../vendor") do
      fetcher.fetch_untar("#{binary_path}.tar.gz")
    end
    FileUtils.ln_s("../vendor/#{binary_path}/bin/node", "node")
    FileUtils.ln_s("../vendor/#{binary_path}/bin/npm", "npm")
  end

  def binary_path
    @binary_path ||= "node-v#{version}-linux-x64"
  end

  private

  def fetcher
    nodejs_base_url = "https://buildpacks.cloudfoundry.org/dependencies/node/v#{version}/"
    @fetcher ||= LanguagePack::Fetcher.new(nodejs_base_url)
  end

  def version
    return @version if @version
    bin_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "compile-extensions", "bin"))
    manifest_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "manifest.yml"))
    @version = `#{bin_path}/default_version_for #{manifest_path} node`.chomp
  end
end
