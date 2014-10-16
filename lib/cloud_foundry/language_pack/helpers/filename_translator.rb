module LanguagePack
  module FilenameTranslator
    DEPENDENCIES_TRANSLATION_REGEX = /[:\/]/
    DEPENDENCIES_TRANSLATION_DELIMITER = '_'

    def self.translate(host_url, original_filename)
      prefix = host_url.to_s.gsub(DEPENDENCIES_TRANSLATION_REGEX, DEPENDENCIES_TRANSLATION_DELIMITER)
      "#{prefix}#{delimiter_for(prefix)}#{original_filename}"
    end

    private

    def self.delimiter_for(prefix)
      (prefix.end_with? '_') ? '' : DEPENDENCIES_TRANSLATION_DELIMITER
    end
  end
end
