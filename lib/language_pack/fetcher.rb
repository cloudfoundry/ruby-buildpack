require "yaml"
require "language_pack/shell_helpers"

module LanguagePack
  class Fetcher
    class FetchError < StandardError; end

    include ShellHelpers
    CDN_YAML_FILE = File.expand_path("../../../config/cdn.yml", __FILE__)

    def initialize(host_url, stack = nil)
      @config   = load_config
      @host_url = fetch_cdn(host_url)
      @host_url += File.basename(stack) if stack
    end

    def fetch(path)
      download_url = @host_url.join(path)
      output_directory = Dir.pwd
      buildpack_dir = File.expand_path(File.join(File.dirname(__FILE__), "..", ".."))
      bin_path = File.join(buildpack_dir, "compile-extensions", "bin")
      download_command = "#{bin_path}/download_dependency #{download_url} #{output_directory}"

      filtered_url = `#{download_command}`.strip

      unless $?.success?
        error_message = `#{bin_path}/recommend_dependency #{download_url}`
        message = "\nCommand: '#{download_command}' failed unexpectedly:\n#{error_message}"
        raise FetchError, message
      end

      system "#{bin_path}/warn_if_newer_patch #{download_url} #{File.join(buildpack_dir, 'manifest.yml')}"

      puts "Downloaded [#{filtered_url}]"
      filtered_url
    end

    def fetch_untar(path, files_to_extract = nil)
      fetch(path)
      tar_command = "tar zxf #{path} #{files_to_extract} && rm #{path}"

      run!(tar_command, error_class: FetchError)
    end

    private
    def curl_timeout_in_seconds
      env('CURL_TIMEOUT') || 90
    end

    def curl_connect_timeout_in_seconds
      env('CURL_CONNECT_TIMEOUT') || 10
    end

    def load_config
      YAML.load_file(CDN_YAML_FILE) || {}
    end

    def fetch_cdn(url)
      url = @config[url] || url
      Pathname.new(url)
    end
  end
end
