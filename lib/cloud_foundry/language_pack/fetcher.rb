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

      if $?.exitstatus != 0
        puts(`#{bin_path}/recommend_dependency #{url}`.chomp)
        exit 1
      end

      filtered_url = `#{bin_path}/filter_dependency_url #{translated_url}`.chomp
      puts "Downloaded [#{filtered_url}]"
      [rendered_command.gsub(url, translated_url), rendered_command.gsub(url,filtered_url)]
    end
  end
end
