require 'cf_spec_helper'

describe "Fetcher" do
  describe "#fetch" do
    context 'with a URL' do
      let(:host_url) { 'scheme://example.com' }

      it 'calls out to translate_dependency_url' do
        fetcher = LanguagePack::Fetcher.new(host_url)
        expect(fetcher).to receive(:`).with(%r{compile-extensions/bin/translate_dependency_url scheme://example.com/ruby.zip}).and_return('scheme://another.com/ruby.zip')
        expect(fetcher).to receive(:run!).with(%r{curl.*scheme://another.com/ruby.zip}, anything)
        fetcher.fetch('ruby.zip')
      end
    end
  end
end

