require_relative '../cf_spec_helper'

describe "OnlineFetcher" do
  let(:path) { "my/path" }

  describe ".fetch_untar" do
    let(:curl_command_callback) { double(:curl_command_callback) }
    let(:run_callback) { double(:run_callback) }
    let(:files_to_extract) { "some_file_to_extract some_other_file_to_extract" }
    let(:host_url) { Pathname.new("http://host.url/") }
    let(:path) { "my/path" }

    it "calls curl and extracts the given files" do
      expect(curl_command_callback).to receive(:call).
        with("http://host.url/my/path -s -o").
        and_return("curl command")
      expect(run_callback).to receive(:call).with("curl command - | tar zxf - some_file_to_extract some_other_file_to_extract")

      OnlineFetcher.fetch_untar(path, host_url, files_to_extract, curl_command_callback, run_callback)
    end
  end
end
