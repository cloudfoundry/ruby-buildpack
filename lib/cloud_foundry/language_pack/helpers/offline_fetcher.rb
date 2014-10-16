class OfflineFetcher
  def self.fetch(original_filename, host_url, error_callback, run_callback)
    dependency_filename = LanguagePack::FilenameTranslator.translate host_url, original_filename

    if DependencyExistenceChecker.exists? dependency_filename
      run_callback.call("cp #{File.join(DEPENDENCIES_PATH, dependency_filename)} #{original_filename}")
    else
      error_callback.call(error_message(original_filename))
    end
  end

  def self.fetch_untar(original_filename, host_url, files_to_extract, error_callback, run_callback)
    dependency_filename = LanguagePack::FilenameTranslator.translate host_url, original_filename

    if DependencyExistenceChecker.exists? dependency_filename
      run_callback.call("tar zxf #{File.join(DEPENDENCIES_PATH, dependency_filename)} #{files_to_extract}")
    else
      error_callback.call(error_message(original_filename))
    end
  end

  private

  def self.error_message original_filename
    "Resource #{original_filename} is not provided by this buildpack. Please upgrade your buildpack to receive the latest resources."
  end
end
