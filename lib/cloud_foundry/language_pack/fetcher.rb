require 'language_pack/fetcher'
require 'cloud_foundry/language_pack/helpers/filename_translator'
require 'cloud_foundry/language_pack/helpers/online_fetcher'
require 'cloud_foundry/language_pack/helpers/offline_fetcher'
require 'cloud_foundry/language_pack/helpers/online_buildpack_detector'
require 'cloud_foundry/language_pack/helpers/dependency_existence_checker'

module LanguagePack
  class Fetcher
    alias_method :heroku_fetch, :fetch

    def fetch(path)
      original_host_url = @host_url
      if requested_mri_version_is_above_212?(path)
        @host_url += 'cedar'
      end
      if OnlineBuildpackDetector.online?
        heroku_fetch path
      else
        OfflineFetcher.fetch(path, @host_url, self.method(:error), self.method(:run!))
      end
    ensure
      @host_url = original_host_url
    end

    def fetch_untar(path, files_to_extract="")
      original_host_url = @host_url
      if requested_mri_version_is_above_212?(path)
        @host_url += 'cedar'
      end
      if OnlineBuildpackDetector.online?
        OnlineFetcher.fetch_untar(path, @host_url, files_to_extract, self.method(:curl_command), self.method(:run!))
      else
        OfflineFetcher.fetch_untar(path, @host_url, files_to_extract, self.method(:error), self.method(:run!))
      end
    ensure
      @host_url = original_host_url
    end

    def requested_mri_version_is_above_212?(path)
      version_string_match = /^ruby-(?<version>[0-9]+\.[0-9]+\.[0-9]+).tgz$/.match(path)
      return false unless version_string_match

      requested_version = Gem::Version.new(version_string_match[:version])
      version_212 = Gem::Version.new("2.1.2")

      requested_version > version_212
    end
  end
end
