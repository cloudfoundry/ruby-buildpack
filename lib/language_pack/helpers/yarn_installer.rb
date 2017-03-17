require_relative '../../../compile-extensions/lib/dependencies'
require 'yaml'

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
    manifest_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "manifest.yml"))
    dependencies = CompileExtensions::Dependencies.new(YAML.load_file(manifest_path))
    @version = dependencies.newest_patch_version({'name'=>'yarn'})
  end
end
