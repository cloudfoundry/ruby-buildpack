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

      it "appends 'cedar' to the host_url if the requested resource is a ruby" do
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
        expect(offline_fetcher).to receive(:fetch).with(path, "#{host_url}cedar", subject.method(:error), subject.method(:run!))
        subject.fetch(path)
      end

      it "appends 'cedar' to the host_url if the requested resource is a ruby" do
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
        expect(online_fetcher).to receive(:fetch_untar).with(path, "#{host_url}cedar", files_to_extract, subject.method(:curl_command), subject.method(:run!))
        subject.fetch_untar(path, files_to_extract)
      end

      it "allows file_to_extract to be passed in optionally" do
        allow(online_fetcher).to receive(:fetch_untar)
        expect{subject.fetch_untar(path)}.not_to raise_error
      end

      it "appends 'cedar' to the host_url if the requested resource is a ruby" do
        expect(online_fetcher).to receive(:fetch_untar).with("ruby-2.1.3.tgz", "#{host_url}cedar", files_to_extract, subject.method(:curl_command), subject.method(:run!))
        subject.fetch_untar("ruby-2.1.3.tgz", files_to_extract)
      end
    end

    context "when there are packaged dependencies (offline buildpack)" do
      let(:is_online) { false }

      it "delegates to offline fetcher" do
        expect(offline_fetcher).to receive(:fetch_untar).with(path, "#{host_url}cedar", files_to_extract, subject.method(:error), subject.method(:run!))
        subject.fetch_untar(path, files_to_extract)
      end

      it "allows file_to_extract to be passed in optionally" do
        allow(offline_fetcher).to receive(:fetch_untar)
        expect{subject.fetch_untar(path)}.not_to raise_error
      end

      it "appends 'cedar' to the host_url if the requested resource is a ruby" do
        expect(offline_fetcher).to receive(:fetch_untar).with("ruby-2.1.3.tgz", "#{host_url}cedar", files_to_extract, subject.method(:error), subject.method(:run!))
        subject.fetch_untar("ruby-2.1.3.tgz", files_to_extract)
      end
    end
  end

  describe "#requested_resource_is_a_ruby?" do
    let(:is_online) { double }

    it "is true for mri" do
      expect(subject.requested_resource_is_a_ruby?("ruby-2.1.3.tgz")).to be_truthy
    end

    it "is true for jruby" do
      expect(subject.requested_resource_is_a_ruby?("ruby-2.0.0-jruby-1.7.11.tgz")).to be_truthy
    end

    it "is truefor ruby-build" do
      expect(subject.requested_resource_is_a_ruby?("ruby-build-1.8.7.tgz")).to be_truthy
    end

    it "is false if the path does not match a ruby package" do
      expect(subject.requested_resource_is_a_ruby?("tomato-potato")).to be_falsy
    end
  end
end

