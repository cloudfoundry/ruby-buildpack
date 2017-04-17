require_relative '../../../compile-extensions/lib/dependencies'
require 'yaml'

class LanguagePack::NodeInstaller
  def initialize(dep_dir, stack)
    @dep_dir = dep_dir
  end

  def install
    Dir.chdir(@dep_dir) do
      fetcher.fetch_untar("#{binary_path}.tar.gz")

      FileUtils.mkdir_p("bin")
      Dir.chdir("bin") do
        FileUtils.ln_s("../#{binary_path}/bin/node", "node")
        FileUtils.ln_s("../#{binary_path}/bin/npm", "npm")
      end
    end
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
    manifest_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "manifest.yml"))
    dependencies = CompileExtensions::Dependencies.new(YAML.load_file(manifest_path))
    if rails51?
      @version = dependencies.newest_patch_version({'name'=>'node', 'version'=>'6.x'})
    else
      @version = dependencies.newest_patch_version({'name'=>'node', 'version'=>'4.x'})
    end
  end

  def rails51?
    bundler = LanguagePack::Helpers::BundlerWrapper.new.install
    rails_version = bundler.gem_version('railties') rescue nil
    rails_version && rails_version >= Gem::Version.new('5.1.0.beta')
  end
end
