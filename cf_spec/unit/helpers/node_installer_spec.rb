require 'cf_spec_helper'

describe LanguagePack::NodeInstaller do
  describe '#install' do
    context 'with different stacks' do
      let(:installer) { LanguagePack::NodeInstaller.new('stack') }

      before do
        allow(FileUtils).to receive(:mv)
        allow(FileUtils).to receive(:rm_rf)
        allow_any_instance_of(LanguagePack::Fetcher).to receive(:fetch_untar)
      end

      it 'always executes the modern Fetcher' do
        expect_any_instance_of(LanguagePack::Fetcher).to receive(:fetch_untar).
          with('node-v4.4.4-linux-x64.tar.gz', 'node-v4.4.4-linux-x64/bin/node')

        installer.install
      end

      it 'moves the node binary to the current path' do
        expect(FileUtils).to receive(:mv).with('node-v4.4.4-linux-x64/bin/node', '.')

        installer.install
      end
    end
  end
end
