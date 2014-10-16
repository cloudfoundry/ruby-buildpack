class DependencyExistenceChecker
  def self.exists?(original_filename)
    File.exists?(File.join(DEPENDENCIES_PATH, original_filename))
  end
end
