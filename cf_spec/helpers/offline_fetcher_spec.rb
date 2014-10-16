require_relative '../cf_spec_helper'

describe "OfflineFetcher" do
  let(:original_filename) { "original_filename.zip" }
  let(:host_url) { Pathname.new("http://host.url/") }
  let(:error_callback) { double(:error_callback) }
  let(:run_callback) { double(:run_callback) }
  let(:dependency_existence_checker) { class_double(DependencyExistenceChecker).as_stubbed_const }
  let(:translated_filename_path) { "http___host.url_original_filename.zip" }

  describe ".fetch" do
    before do
      stub_const("DEPENDENCIES_PATH", "some/dependencies/path")
      allow(dependency_existence_checker).to receive(:exists?).with(translated_filename_path) { exists }
    end

    context "when the dependency exists" do
      let(:exists) { true }

      it "calls run with a copy shell command" do
        expect(run_callback).to receive(:call).with("cp some/dependencies/path/http___host.url_original_filename.zip original_filename.zip")
        OfflineFetcher.fetch(original_filename, host_url, error_callback, run_callback)
      end
    end

    context "when the dependency is missing" do
      let(:exists) { false }

      it "calls error with a friendly error message" do
        expect(error_callback).to receive(:call).with("Resource original_filename.zip is not provided by this buildpack. Please upgrade your buildpack to receive the latest resources.")
        OfflineFetcher.fetch(original_filename, host_url, error_callback, run_callback)
      end
    end
  end

  describe ".fetch_untar" do
    let(:files_to_extract) { "some_file_to_extract some_other_file_to_extract" }

    before do
      stub_const("DEPENDENCIES_PATH", "some/dependencies/path")
      allow(dependency_existence_checker).to receive(:exists?).with(translated_filename_path) { exists }
    end

    context "when the dependency exists" do
      let(:exists) { true }

      it "calls run with a tar shell command" do
        expect(run_callback).to receive(:call).with("tar zxf some/dependencies/path/http___host.url_original_filename.zip some_file_to_extract some_other_file_to_extract")
        OfflineFetcher.fetch_untar(original_filename, host_url, files_to_extract, error_callback, run_callback)
      end
    end

    context "when the dependency is missing" do
      let(:exists) { false }

      it "calls error with a friendly error message" do
        expect(error_callback).to receive(:call).with("Resource original_filename.zip is not provided by this buildpack. Please upgrade your buildpack to receive the latest resources.")
        OfflineFetcher.fetch_untar(original_filename, host_url, files_to_extract, error_callback, run_callback)
      end
    end
  end
end


