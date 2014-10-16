require_relative '../cf_spec_helper'
require 'cloud_foundry/language_pack/helpers/filename_translator'

describe "LanguagePack::FilenameTranslator" do
  describe ".translate" do
    it "returns a file name prefixed with the encoded host url" do
      expect(LanguagePack::FilenameTranslator.translate("http://google.com", "file.zip")).to eq "http___google.com_file.zip"
    end
  end
end
