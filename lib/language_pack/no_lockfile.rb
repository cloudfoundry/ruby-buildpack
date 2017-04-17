require "language_pack"
require "language_pack/base"

class LanguagePack::NoLockfile < LanguagePack::Base
  def self.use?
    !File.exists?("Gemfile.lock")
  end

  def name
    "Ruby/NoLockfile"
  end

  def supply
    error "Gemfile.lock required. Please check it in."
  end

  def finalize
    error "Gemfile.lock required. Please check it in."
  end
end
