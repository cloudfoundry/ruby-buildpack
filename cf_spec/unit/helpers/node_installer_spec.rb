require 'cf_spec_helper'

describe LanguagePack::NodeInstaller do
  describe '#install' do
    context 'with different stacks' do
      let(:dep_dir) { Dir.mktmpdir }
      let(:installer) { LanguagePack::NodeInstaller.new(dep_dir, 'stack') }
      let(:railsversion) { Gem::Version.new('4.2.0') }
      let(:bundler) { double(:bundler_wrapper) }

      before do
        allow(Dir).to receive(:chdir).and_yield
        allow(FileUtils).to receive(:ln_s)
        allow_any_instance_of(LanguagePack::Helpers::BundlerWrapper).to receive(:install).and_return(bundler)
        allow_any_instance_of(LanguagePack::Fetcher).to receive(:fetch_untar)
        allow(bundler).to receive(:gem_version).and_return(railsversion)
      end

      after do
        FileUtils.rm_rf(dep_dir)
      end

      it 'executes the fetcher' do
        node_tar_regex = Regexp.new('node\-v4\.\d+\.\d+' + Regexp.quote('-linux-x64.tar.gz'))
        expect_any_instance_of(LanguagePack::Fetcher).to receive(:fetch_untar).with(node_tar_regex)
        installer.install
      end

      it 'links the node binary to bin' do
        extract_path_regex = Regexp.new('node\-v4\.\d+\.\d+' + Regexp.quote('-linux-x64/bin/node'))
        expect(FileUtils).to receive(:ln_s).with(extract_path_regex, 'node')
        installer.install
      end

      it 'links the npm binary to bin' do
        extract_path_regex = Regexp.new('node\-v4\.\d+\.\d+' + Regexp.quote('-linux-x64/bin/npm'))
        expect(FileUtils).to receive(:ln_s).with(extract_path_regex, 'npm')
        installer.install
      end

      context 'rails version 5.1.0' do
        let(:railsversion) {Gem::Version.new('5.1.0')}

        it 'downloads node 6' do
          node_tar_regex = Regexp.new('node\-v6\.\d+\.\d+' + Regexp.quote('-linux-x64.tar.gz'))
          expect_any_instance_of(LanguagePack::Fetcher).to receive(:fetch_untar).with(node_tar_regex)
          installer.install
        end
      end
    end
  end
end
