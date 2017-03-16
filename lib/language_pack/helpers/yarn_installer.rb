class LanguagePack::YarnInstaller
  def initialize(stack)
  end

  def install
    Dir.chdir("../vendor") do
      FileUtils.mkdir_p(binary_path)
      Dir.chdir(binary_path) do
        fetcher.fetch_untar("#{binary_path}.tar.gz", "--strip-components 1")
      end
    end

    FileUtils.ln_s("../vendor/#{binary_path}/bin/yarn", "yarn", :force => true)
  end

  def binary_path
    @binary_path || "yarn-v#{version}"
  end

  private

  def fetcher
    yarn_base_url = "https://yarnpkg.com/downloads/#{version}/#{binary_path}.tar.gz"
    @fetcher ||= LanguagePack::Fetcher.new(yarn_base_url)
  end

  def version
    return @version if @version
    bin_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "compile-extensions", "bin"))
    manifest_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "manifest.yml"))
    @version = `#{bin_path}/default_version_for #{manifest_path} yarn`.chomp
  end
end
