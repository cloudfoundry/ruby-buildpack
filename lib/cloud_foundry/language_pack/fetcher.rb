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
      if requested_resource_is_a_ruby?(path)
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
      if requested_resource_is_a_ruby?(path)
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

    def requested_resource_is_a_ruby?(path)
      return (path.match /^ruby/) ? true : false
    end
  end
end
