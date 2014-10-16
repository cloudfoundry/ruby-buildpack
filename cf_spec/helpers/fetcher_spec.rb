require_relative '../cf_spec_helper'

describe "Fetcher" do
  subject { LanguagePack::Fetcher.new(host_url) }

  let(:detector) { class_double("OnlineBuildpackDetector").as_stubbed_const }
  let(:online_fetcher) { class_double("OnlineFetcher").as_stubbed_const }
  let(:offline_fetcher) { class_double("OfflineFetcher").as_stubbed_const }
  let(:host_url) { "localhost" }
  let(:path) { double(:path) }
  let(:pathname) { class_double("Pathname").as_stubbed_const }

  before do
    allow(pathname).to receive(:new).and_return(host_url)
    allow(detector).to receive(:online?).and_return(is_online)
  end

  describe "#fetch" do
    context "when there are no packaged dependencies (online buildpack)" do
      let(:is_online) { true }

      it "delegates to heroku's version of fetch" do
        expect(subject).to receive(:heroku_fetch).with(path)
        subject.fetch(path)
      end
    end

    context "when there are packaged dependencies (offline buildpack)" do
      let(:is_online) { false }

      it "delegates to offline fetcher" do
        expect(offline_fetcher).to receive(:fetch).with(path, host_url, subject.method(:error), subject.method(:run!))
        subject.fetch(path)
      end
    end
  end

  describe "#fetch_untar" do
    let(:files_to_extract) { double(:files_to_extract) }


    context "when there are no packaged dependencies (online buildpack)" do
      let(:is_online) { true }

      it "delegates to online fetcher" do
        expect(online_fetcher).to receive(:fetch_untar).with(path, host_url, files_to_extract, subject.method(:curl_command), subject.method(:run!))
        subject.fetch_untar(path, files_to_extract)
      end

      it "allows file_to_extract to be passed in optionally" do
        allow(online_fetcher).to receive(:fetch_untar)
        expect{subject.fetch_untar(path)}.not_to raise_error
      end
    end

    context "when there are packaged dependencies (offline buildpack)" do
      let(:is_online) { false }

      it "delegates to offline fetcher" do
        expect(offline_fetcher).to receive(:fetch_untar).with(path, host_url, files_to_extract, subject.method(:error), subject.method(:run!))
        subject.fetch_untar(path, files_to_extract)
      end

      it "allows file_to_extract to be passed in optionally" do
        allow(offline_fetcher).to receive(:fetch_untar)
        expect{subject.fetch_untar(path)}.not_to raise_error
      end
    end
  end
end

