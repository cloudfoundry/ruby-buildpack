module LanguagePack
  module BundlerLockfile
    module ClassMethods
      def bundle_gemfile
        @bundle_gemfile ||= ENV["BUNDLE_GEMFILE"] || "Gemfile"
      end

      # checks if the Gemfile and Gemfile.lock exist
      def gemfile_lock?
        File.exist?(bundle_gemfile) && File.exist?("#{bundle_gemfile}.lock")
      end

      def bundle
        @bundle ||= parse_bundle
      end

      def bundler_path
        @bundler_path ||= fetch_bundler
      end

      def fetch_bundler
        Dir.mktmpdir("bundler-").tap do |dir|
          Dir.chdir(dir) do
            fetch_package_and_untar("#{LanguagePack::Ruby::BUNDLER_GEM_PATH}.tgz")
          end
        end
      end

      def parse_bundle
        $: << "#{bundler_path}/gems/bundler-#{LanguagePack::Ruby::BUNDLER_VERSION}/lib"
        require "bundler"
        Bundler::LockfileParser.new(File.read("#{bundle_gemfile}.lock"))
      end
    end

    def bundle
      self.class.bundle
    end

    def bundle_gemfile
      self.class.bundle_gemfile
    end

    def bundler_path
      self.class.bundler_path
    end
  end
end
