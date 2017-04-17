require "language_pack/shell_helpers"

class LanguagePack::JvmInstaller
  include LanguagePack::ShellHelpers

  SYS_PROPS_FILE  = "system.properties"
  JVM_BUCKET      = "https://lang-jvm.s3.amazonaws.com"
  JVM_BASE_URL    = "#{JVM_BUCKET}/jdk"
  JVM_1_8_PATH    = "openjdk1.8-latest.tar.gz"

  def initialize(dep_dir, stack)
    @dep_dir = dep_dir
    @vendor_dir = "#{dep_dir}/jvm"
    @stack = stack
    @fetcher = LanguagePack::Fetcher.new(JVM_BASE_URL, stack)
  end

  def system_properties
    props = {}
    File.read(SYS_PROPS_FILE).split("\n").each do |line|
      key = line.split("=").first
      val = line.split("=").last
      props[key] = val
    end if File.exists?(SYS_PROPS_FILE)
    props
  end

  def install(jruby_version, forced = false)
    if Dir.exist?(".jdk")
      topic "Using pre-installed JDK"
      return
    end

    fetch_untar(JVM_1_8_PATH, "openjdk-8")

    FileUtils.mkdir_p "#{@dep_dir}/bin"
    Dir.chdir("#{@dep_dir}/bin") do
      Dir["../jvm/bin/*"].each do |bin|
        FileUtils.ln_s(bin, File.basename(bin))
      end
    end
  end

  def fetch_untar(jvm_path, jvm_version=nil)
    topic "Installing JVM: #{jvm_version || jvm_path}"

    FileUtils.mkdir_p(@vendor_dir)
    Dir.chdir(@vendor_dir) do
      @fetcher.fetch_untar(jvm_path)
    end
  rescue LanguagePack::Fetcher::FetchError
    error <<EOF
Failed to download JVM: #{jvm_path}

If this was a custom version or URL, please check to ensure it is correct.
EOF
  end
end
