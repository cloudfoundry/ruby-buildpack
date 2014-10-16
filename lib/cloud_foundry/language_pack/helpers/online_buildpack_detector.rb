class OnlineBuildpackDetector
  def self.online?
    !Dir.exist? DEPENDENCIES_PATH
  end
end
