require 'cf_spec_helper'

describe LanguagePack::NodeInstaller do
  describe '#install' do
    context 'with different stacks' do
      let(:installer) { LanguagePack::NodeInstaller.new('stack') }

      before do
        allow(Dir).to receive(:chdir).and_yield
        allow(FileUtils).to receive(:ln_s)
        allow_any_instance_of(LanguagePack::Fetcher).to receive(:fetch_untar)
      end

      it 'executes the fetcher' do
        node_tar_regex = Regexp.new('node\-v\d+\.\d+\.\d+' + Regexp.quote('-linux-x64.tar.gz'))
        expect_any_instance_of(LanguagePack::Fetcher).to receive(:fetch_untar).with(node_tar_regex)
        installer.install
      end

      it 'links the node binary to bin' do
        extract_path_regex = Regexp.new('node\-v\d+\.\d+\.\d+' + Regexp.quote('-linux-x64/bin/node'))
        expect(FileUtils).to receive(:ln_s).with(extract_path_regex, 'node')
        installer.install
      end

      it 'links the npm binary to bin' do
        extract_path_regex = Regexp.new('node\-v\d+\.\d+\.\d+' + Regexp.quote('-linux-x64/bin/npm'))
        expect(FileUtils).to receive(:ln_s).with(extract_path_regex, 'npm')
        installer.install
      end
    end
  end
end
