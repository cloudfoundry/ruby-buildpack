require 'language_pack/fetcher'
require 'uri'

module LanguagePack
  class Fetcher
    alias_method :original_curl_command, :curl_command

    private

    def curl_command(command)
      rendered_command = original_curl_command(command)
      url = rendered_command.match(URI.regexp)[0]
      bin_path = File.expand_path(File.join(File.dirname(__FILE__), "..", "..", "..", "compile-extensions", "bin"))
      translated_url = `#{bin_path}/translate_dependency_url #{url}`.chomp
      rendered_command.sub(url, translated_url)
    end
  end
end