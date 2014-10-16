class OnlineFetcher
  def self.fetch_untar(path, host_url, files_to_extract, curl_command_callback, run_callback)
    curl_command = curl_command_callback.call("#{host_url.join(path)} -s -o")
    run_callback.call("#{curl_command} - | tar zxf - #{files_to_extract}")
  end
end
