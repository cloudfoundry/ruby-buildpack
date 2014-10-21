require_relative '../cf_spec_helper'

describe "Fetcher" do
  subject { LanguagePack::Fetcher.new(host_url) }

  let(:detector) { class_double("OnlineBuildpackDetector").as_stubbed_const }
  let(:online_fetcher) { class_double("OnlineFetcher").as_stubbed_const }
  let(:offline_fetcher) { class_double("OfflineFetcher").as_stubbed_const }
  let(:host_url) { "localhost/" }
  let(:path) { "ruby-2.1.2.tgz" }
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

      it "appends 'cedar' to the host_url if the requested resource is ruby 2.1.3 or above" do
        expect(subject).to receive(:heroku_fetch) do |path|
          expect(path).to eq("ruby-2.1.3.tgz")
          expect(subject.instance_variable_get(:@host_url)).to include "/cedar"
        end
        subject.fetch("ruby-2.1.3.tgz")
      end
    end

    context "when there are packaged dependencies (offline buildpack)" do
      let(:is_online) { false }

      it "delegates to offline fetcher" do
        expect(offline_fetcher).to receive(:fetch).with(path, host_url, subject.method(:error), subject.method(:run!))
        subject.fetch(path)
      end

      it "appends 'cedar' to the host_url if the requested resource is ruby 2.1.3 or above" do
        expect(offline_fetcher).to receive(:fetch).with("ruby-2.1.3.tgz", "#{host_url}cedar", subject.method(:error), subject.method(:run!))
        subject.fetch("ruby-2.1.3.tgz")
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

      it "appends 'cedar' to the host_url if the requested resource is ruby 2.1.3 or above" do
        expect(online_fetcher).to receive(:fetch_untar).with("ruby-2.1.3.tgz", "#{host_url}cedar", files_to_extract, subject.method(:curl_command), subject.method(:run!))
        subject.fetch_untar("ruby-2.1.3.tgz", files_to_extract)
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

      it "appends 'cedar' to the host_url if the requested resource is ruby 2.1.3 or above" do
        expect(offline_fetcher).to receive(:fetch_untar).with("ruby-2.1.3.tgz", "#{host_url}cedar", files_to_extract, subject.method(:error), subject.method(:run!))
        subject.fetch_untar("ruby-2.1.3.tgz", files_to_extract)
      end
    end
  end

  describe "#requested_ruby_version_is_above_212?" do
    let(:is_online) { double }

    it "is true if a ruby version higher than 2.1.2 is found in the path" do
      expect(subject.requested_ruby_version_is_above_212?("ruby-2.1.3.tgz")).to be_truthy
      expect(subject.requested_ruby_version_is_above_212?("ruby-2.1.4.tgz")).to be_truthy
      expect(subject.requested_ruby_version_is_above_212?("ruby-2.2.0.tgz")).to be_truthy
      expect(subject.requested_ruby_version_is_above_212?("ruby-2.1.11.tgz")).to be_truthy
    end

    it "is false if a ruby version equal to 2.1.2 is found in the path" do
      expect(subject.requested_ruby_version_is_above_212?("ruby-2.1.2.tgz")).to be_falsy
    end

    it "is false if a ruby version below 2.1.2 is found in the path" do
      expect(subject.requested_ruby_version_is_above_212?("ruby-1.9.3.tgz")).to be_falsy
    end

    it "is false if the path does not match a ruby package" do
      expect(subject.requested_ruby_version_is_above_212?("tomato-potato")).to be_falsy
    end
  end
end

