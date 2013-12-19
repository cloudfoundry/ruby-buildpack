require "language_pack"
require "language_pack/base"

class LanguagePack::NoLockfile < LanguagePack::Base
  def self.bundle_gemfile
    @bundle_gemfile ||= ENV["BUNDLE_GEMFILE"] || "Gemfile"
  end

  def self.use?
    File.exist?(bundle_gemfile) && !File.exists?("#{bundle_gemfile}.lock")
  end

  def name
    "Ruby/NoLockfile"
  end

  def compile
    error "#{bundle_gemfile}.lock required. Please check it in."
  end
end
