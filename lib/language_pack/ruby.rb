require "tmpdir"
require "digest/md5"
require "benchmark"
require "rubygems"
require "language_pack"
require "language_pack/base"
require "language_pack/ruby_version"
require "language_pack/helpers/node_installer"
require "language_pack/helpers/yarn_installer"
require "language_pack/helpers/jvm_installer"
require "language_pack/version"
require "language_pack/default_version"

# base Ruby Language Pack. This is for any base ruby app.
class LanguagePack::Ruby < LanguagePack::Base
  NAME                 = "ruby"
  BUNDLER_VERSION      = LanguagePack::DefaultVersion.for('bundler')
  BUNDLER_GEM_PATH     = "bundler-#{BUNDLER_VERSION}"
  NODE_BP_PATH         = "vendor/node/bin"

  # detects if this is a valid Ruby app
  # @return [Boolean] true if it's a Ruby app
  def self.use?
    instrument "ruby.use" do
      File.exist?("Gemfile")
    end
  end

  def self.bundler
    @@bundler ||= LanguagePack::Helpers::BundlerWrapper.new.install
  end

  def bundler
    self.class.bundler
  end

  def initialize(build_path, cache_path=nil, dep_dir = "vendor")
    super(build_path, cache_path, dep_dir)
    @fetchers[:mri]    = LanguagePack::Fetcher.new(VENDOR_URL, @stack)
    @node_installer    = LanguagePack::NodeInstaller.new(dep_dir, @stack)
    @yarn_installer    = LanguagePack::YarnInstaller.new(dep_dir, @stack)
    @jvm_installer     = LanguagePack::JvmInstaller.new(dep_dir, @stack)
  end

  def name
    "Ruby"
  end

  def default_config_vars
    instrument "ruby.default_config_vars" do
      vars = {
        "LANG" => env("LANG") || "en_US.UTF-8"
      }

      ruby_version.jruby? ? vars.merge({
        "JAVA_OPTS" => default_java_opts,
        "JRUBY_OPTS" => default_jruby_opts
      }) : vars
    end
  end

  def default_process_types
    instrument "ruby.default_process_types" do
      {
        "rake"    => "bundle exec rake",
        "console" => "bundle exec irb"
      }
    end
  end

  def best_practice_warnings; end

  def supply
    instrument 'ruby.supply' do
      # check for new app at the beginning of supply
      new_app?
      Dir.chdir(build_path)
      warn_bundler_upgrade
      warn_windows_gemfile_endline
      install_ruby
      install_jvm
      install_binaries
      add_dep_dir_to_path
      setup_export
      setup_profiled
    end
  end

  def finalize
    instrument 'ruby.finalize' do
      # check for new app at the beginning of finalize
      new_app?
      Dir.chdir(build_path)
      remove_vendor_bundle
      link_supplied_binaries_in_app
      allow_git do
        install_bundler_in_app
        build_bundler
        post_bundler
        create_database_yml
        run_assets_precompile_rake_task
      end
      best_practice_warnings
      super
    end
  end

private

  def warn_bundler_upgrade
    old_bundler_version  = @metadata.read("bundler_version").chomp if @metadata.exists?("bundler_version")

    if old_bundler_version && old_bundler_version != BUNDLER_VERSION
      puts(<<-WARNING)
Your app was upgraded to bundler #{ BUNDLER_VERSION }.
Previously you had a successful deploy with bundler #{ old_bundler_version }.

WARNING
    end
  end

  def warn_windows_gemfile_endline
    filename = ENV.fetch("BUNDLE_GEMFILE", "Gemfile")
    file_contents = File.read(filename)
    if file_contents.include?("\r\n")
      puts("WARNING: Windows line endings detected in Gemfile. Your app may fail to stage. Please use UNIX line endings.")
    end
  end

  def binstubs_relative_paths
    [
      "bin",
      bundler_binstubs_path,
      "#{slug_vendor_base}/bin"
    ]
  end

  # the relative path to the bundler directory of gems
  # @return [String] resulting path
  def slug_vendor_base
    instrument 'ruby.slug_vendor_base' do
      return @slug_vendor_base if @slug_vendor_base

      @slug_vendor_base = run_no_pipe(%q(ruby -e "require 'rbconfig';puts \"vendor/bundle/#{RUBY_ENGINE}/#{RbConfig::CONFIG['ruby_version']}\"")).chomp
      error "Problem detecting bundler vendor directory: #{@slug_vendor_base}" unless $?.success?
      @slug_vendor_base
    end
  end

  # the relative path to the vendored ruby directory
  # @return [String] resulting path
  def slug_vendor_ruby
    "#{@dep_dir}/#{ruby_version.version_without_patchlevel}"
  end

  # the relative path to the vendored jvm
  # @return [String] resulting path
  def slug_vendor_jvm
    "#{@dep_dir}/jvm"
  end

  # fetch the ruby version from bundler
  # @return [String, nil] returns the ruby version if detected or nil if none is detected
  def ruby_version
    instrument 'ruby.ruby_version' do
      return @ruby_version if @ruby_version
      new_app           = !File.exist?("vendor/.cloudfoundry/metadata")
      last_version_file = "buildpack_ruby_version"
      last_version      = nil
      last_version      = @metadata.read(last_version_file).chomp if @metadata.exists?(last_version_file)

      @ruby_version = LanguagePack::RubyVersion.new(bundler.ruby_version,
        is_new:       new_app,
        last_version: last_version)
      return @ruby_version
    end
  end

  # default JAVA_OPTS
  # return [String] string of JAVA_OPTS
  def default_java_opts
    "-Xss512k -XX:+UseCompressedOops -Dfile.encoding=UTF-8"
  end

  def staging_jvm_max_heap
    ulimit = `bash -c 'ulimit -u'`.strip

    case ulimit
      when "256" then "384"
      when "512" then "768"
      when "16384" then "2048"
      when "32768" then "5120"
      else ""
    end
  end

  # staging Java Xmx
  # return [String] string of Java Xmx
  def staging_java_mem
    max_heap = staging_jvm_max_heap

    if max_heap != ""
      "-Xmx#{max_heap}m"
    else
      "-Xmx384m"
    end
  end

  def runtime_jvm_max_heap
    <<-EOF
case $(ulimit -u) in
256)   # 1X Dyno
  JVM_MAX_HEAP=384
  ;;
512)   # 2X Dyno
  JVM_MAX_HEAP=768
  ;;
16384) # IX Dyno
  JVM_MAX_HEAP=2048
  ;;
32768) # PX Dyno
  JVM_MAX_HEAP=5120
  ;;
esac
EOF
  end

  def runtime_java_mem
    <<-EOF
if ! [[ "${JAVA_OPTS}" == *-Xmx* ]]; then
  export JAVA_MEM=${JAVA_MEM:--Xmx${JVM_MAX_HEAP:-384}m}
fi
EOF
  end

  def set_default_web_concurrency
    <<-EOF
case $(ulimit -u) in
256)
  export HEROKU_RAM_LIMIT_MB=${HEROKU_RAM_LIMIT_MB:-512}
  export WEB_CONCURRENCY=${WEB_CONCURRENCY:-2}
  ;;
512)
  export HEROKU_RAM_LIMIT_MB=${HEROKU_RAM_LIMIT_MB:-1024}
  export WEB_CONCURRENCY=${WEB_CONCURRENCY:-4}
  ;;
16384)
  export HEROKU_RAM_LIMIT_MB=${HEROKU_RAM_LIMIT_MB:-2560}
  export WEB_CONCURRENCY=${WEB_CONCURRENCY:-8}
  ;;
32768)
  export HEROKU_RAM_LIMIT_MB=${HEROKU_RAM_LIMIT_MB:-6144}
  export WEB_CONCURRENCY=${WEB_CONCURRENCY:-16}
  ;;
*)
  ;;
esac
EOF
  end

  # default JRUBY_OPTS
  # return [String] string of JRUBY_OPTS
  def default_jruby_opts
    "-Xcompile.invokedynamic=false"
  end


  # we need this so supply and finalize use the same ruby when they call slug_vendor_base
  def add_dep_dir_to_path
    ENV['PATH'] = "#{@dep_dir}/bin:#{ENV['PATH']}"
    if ENV['GEM_PATH']
      ENV['GEM_PATH'] = "#{@dep_dir}/bundler-#{BUNDLER_VERSION}:#{ENV['GEM_PATH']}"
    else
      ENV['GEM_PATH'] = "#{@dep_dir}/bundler-#{BUNDLER_VERSION}"
    end
  end

  # Sets up the environment variables for subsequent processes run by
  # muiltibuildpack. We can't use profile.d because $HOME isn't set up
  def setup_export
    instrument 'ruby.setup_export' do
      write_env_file "GEM_PATH", "#{build_path}/#{slug_vendor_base}:#{ENV['GEM_PATH']}"
      write_env_file "GEM_HOME", "#{build_path}/#{slug_vendor_base}"
      write_env_file "LANG",     "en_US.UTF-8"

      config_vars = default_config_vars.each do |key, value|
        write_env_file key, value
      end

      if ruby_version.jruby?
        puts "Using Java Memory: #{staging_java_mem}"
        write_env_file "JVM_MAX_HEAP", staging_jvm_max_heap if staging_jvm_max_heap != ""
        write_env_file "JAVA_MEM",   staging_java_mem
        write_env_file "JAVA_OPTS",  default_java_opts
        write_env_file "JRUBY_OPTS", default_jruby_opts
      end
    end
  end

  # sets up the profile.d script for this buildpack
  def setup_profiled
    instrument 'setup_profiled' do
      set_env_default  "LANG",     "en_US.UTF-8"
      dep_idx = File.basename(@dep_dir)
      add_to_profiled %{export GEM_PATH="$HOME/#{slug_vendor_base}:$DEPS_DIR/#{dep_idx}/bundler-#{BUNDLER_VERSION}$([[ ! -z "${GEM_PATH:-}" ]] && echo ":$GEM_PATH")"}
      set_env_override "PATH",     binstubs_relative_paths.map {|path| "$HOME/#{path}" }.join(":") + ":$PATH"

      add_to_profiled set_default_web_concurrency if env("SENSIBLE_DEFAULTS")

      if ruby_version.jruby?
        add_to_profiled runtime_jvm_max_heap
        add_to_profiled runtime_java_mem
        set_env_default "JAVA_OPTS", default_java_opts
        set_env_default "JRUBY_OPTS", default_jruby_opts
      end

      add_to_profiled <<-HERE
if [[ ! -z "${LD_LIBRARY_PATH:-}" ]]; then
 export LD_LIBRARY_PATH="$HOME/ld_library_path:$LD_LIBRARY_PATH"
else
 export LD_LIBRARY_PATH="$HOME/ld_library_path"
fi
HERE
    end
  end

  # install the vendored ruby
  # @return [Boolean] true if it installs the vendored ruby and false otherwise
  def install_ruby
    instrument 'ruby.install_ruby' do
      return false unless ruby_version

      FileUtils.mkdir_p(slug_vendor_ruby)
      Dir.chdir(slug_vendor_ruby) do
        instrument "ruby.fetch_ruby" do
          if ruby_version.rbx?
            error(<<-ERROR)
Rubinius is not supported by this buildpack, please choose a different engine
ERROR

          else
            @fetchers[:mri].fetch_untar("#{ruby_version.version_for_download}.tgz")
          end
        end
      end

      ## Change rake hashbang
      if File.exists?("#{slug_vendor_ruby}/bin/rake")
        rake_contents = File.read("#{slug_vendor_ruby}/bin/rake")
        if rake_contents.gsub!(%r{/app/vendor/.*/bin/ruby}, '/usr/bin/env ruby')
          File.write("#{slug_vendor_ruby}/bin/rake", rake_contents)
        end
      end

      FileUtils.ln_s("ruby", "#{slug_vendor_ruby}/bin/ruby.exe")
      dest = Pathname.new("#{@dep_dir}/bin")
      FileUtils.mkdir_p(dest.to_s)
      Dir["#{slug_vendor_ruby}/bin/*"].each do |bin|
        relative_bin = Pathname.new(bin).relative_path_from(dest).to_s
        FileUtils.ln_s(relative_bin, "#{dest}/#{File.basename(bin)}")
      end

      @metadata.write("buildpack_ruby_version", ruby_version.version_for_download)

      topic "Using Ruby version: #{ruby_version.version_for_download}"
      if !ruby_version.set
        warn(<<-WARNING)
You have not declared a Ruby version in your Gemfile.
To set your Ruby version add this line to your Gemfile:
#{ruby_version.to_gemfile}
# See http://docs.cloudfoundry.org/buildpacks/ruby/index.html#runtime for more information.
WARNING
      end
    end

    true
  rescue LanguagePack::Fetcher::FetchError => error
    message = <<ERROR
An error occurred while installing #{ruby_version.version_for_download}

It is recommended you use the latest supported Ruby version listed here:
  http://docs.cloudfoundry.org/buildpacks/ruby/#supported_versions

For more information on syntax for declaring a Ruby version see:
  http://docs.cloudfoundry.org/buildpacks/ruby/index.html#runtime

ERROR
    message << "\nDebug Information"
    message << error.message

    error message
  end

  def link_supplied_binaries_in_app
    dep_idx = File.basename(@dep_dir)
    dest = Pathname.new("#{build_path}/bin")
    FileUtils.mkdir_p(dest.to_s)
    Dir["#{@dep_dir}/bin/*"].each do |bin|
      dest_file = "#{dest}/#{File.basename(bin)}"
      unless File.exists?(dest_file)
        File.write(dest_file, %Q{#!/bin/bash\n$DEPS_DIR/#{dep_idx}/bin/#{File.basename(bin)} "$@"\n})
        FileUtils.chmod '+x', dest_file
      end
    end
  end

  def new_app?
    @new_app ||= !File.exist?("vendor/.cloudfoundry/metadata/stack")
  end

  # vendors JVM into the slug for JRuby
  def install_jvm(forced = false)
    instrument 'ruby.install_jvm' do
      if ruby_version.jruby? || forced
        @jvm_installer.install(ruby_version.engine_version, forced)
      end
    end
  end

  # installs vendored gems into the slug
  def install_bundler_in_app
    instrument 'ruby.install_language_pack_gems' do
      FileUtils.mkdir_p(slug_vendor_base)
      Dir.chdir(slug_vendor_base) do |dir|
        `cp -R #{bundler.bundler_path}/. .`
      end
    end
  end

  # default set of binaries to install
  # @return [Array] resulting list
  def binaries
    bins = add_node_js_binary
    if File.exist?("yarn.lock")
      bins << 'yarn'
    end
    bins
  end

  # vendors binaries into the slug
  def install_binaries
    instrument 'ruby.install_binaries' do
      binaries.each {|binary| install_binary(binary) }
      Dir["bin/*"].each {|path| run("chmod +x #{path}") }
    end
  end

  # vendors individual binary into the slug
  # @param [String] name of the binary package from S3.
  #   Example: https://s3.amazonaws.com/language-pack-ruby/node-0.4.7.tgz, where name is "node-0.4.7"
  def install_binary(name)
    bin_dir = "bin"
    FileUtils.mkdir_p bin_dir
    Dir.chdir(bin_dir) do |dir|
      if name.match(/^node\-/)
        @node_installer.install
      elsif name == 'yarn'
        @yarn_installer.install
      else
        @fetchers[:buildpack].fetch_untar("#{name}.tgz")
      end
    end
  end

  # remove `vendor/bundle` that comes from the git repo
  # in case there are native ext.
  # users should be using `bundle pack` instead.
  def remove_vendor_bundle
    if File.exists?("vendor/bundle")
      warn(<<-WARNING)
Removing `vendor/bundle`.
Checking in `vendor/bundle` is not supported. Please remove this directory
and add it to your .gitignore. To vendor your gems with Bundler, use
`bundle pack` instead.
WARNING
      FileUtils.rm_rf("vendor/bundle")
    end
  end

  def bundler_binstubs_path
    "vendor/bundle/bin"
  end

  # runs bundler to install the dependencies
  def build_bundler
    instrument 'ruby.build_bundler' do
      log("bundle") do
        bundle_without = env("BUNDLE_WITHOUT") || "development:test"
        bundle_bin     = "bundle"
        bundle_command = "#{bundle_bin} install --without #{bundle_without} --path vendor/bundle --binstubs #{bundler_binstubs_path}"
        bundle_command << " --jobs=4"
        bundle_command << " --retry=4"

        if File.exist?("#{Dir.pwd}/.bundle/config")
          warn(<<-WARNING)
You have the `.bundle/config` file checked into your repository
 It contains local state like the location of the installed bundle
 as well as configured git local gems, and other settings that should
not be shared between multiple checkouts of a single repo. Please
remove the `.bundle/` folder from your repo and add it to your `.gitignore` file.
WARNING
        end

        if bundler.windows_gemfile_lock?
          warn(<<-WARNING)
Removing `Gemfile.lock` because it was generated on Windows.
Bundler will do a full resolve so native gems are handled properly.
This may result in unexpected gem versions being used in your app.
In rare occasions Bundler may not be able to resolve your dependencies at all.
https://docs.cloudfoundry.org/buildpacks/ruby/windows.html
WARNING

          log("bundle", "has_windows_gemfile_lock")
          File.unlink("Gemfile.lock")
        else
          # using --deployment is preferred if we can
          bundle_command += " --deployment"
        end

        topic("Installing dependencies using bundler #{bundler.version}")
        load_bundler_cache

        bundler_output = ""
        bundle_time    = nil

        # need to setup compile environment for the psych gem
        pwd            = Dir.pwd
        bundler_path   = "#{pwd}/#{slug_vendor_base}/gems/#{BUNDLER_GEM_PATH}/lib"
        # we need to set BUNDLE_CONFIG and BUNDLE_GEMFILE for
        # codon since it uses bundler.
        env_vars       = {
          "BUNDLE_GEMFILE"                => "#{pwd}/#{ENV['BUNDLE_GEMFILE']}",
          "BUNDLE_CONFIG"                 => "#{pwd}/.bundle/config",
          "NOKOGIRI_USE_SYSTEM_LIBRARIES" => "true"
        }
        puts "Running: #{bundle_command}"
        instrument "ruby.bundle_install" do
          bundle_time = Benchmark.realtime do
            bundler_output << pipe("#{bundle_command} --no-clean", out: "2>&1", env: env_vars, user_env: true)
          end
        end

        if $?.success?
          puts "Bundle completed (#{"%.2f" % bundle_time}s)"
          log "bundle", :status => "success"
          puts "Cleaning up the bundler cache."
          instrument "ruby.bundle_clean" do
            pipe("#{bundle_bin} clean", out: "2> /dev/null", user_env: true)
          end
          @bundler_cache.store

          # Keep gem cache out of the slug
          FileUtils.rm_rf("#{slug_vendor_base}/cache")
        else
          log "bundle", :status => "failure"
          error_message = "Failed to install gems via Bundler."
          puts "Bundler Output: #{bundler_output}"

          error error_message
        end
      end
    end
  end

  def post_bundler
    instrument "ruby.post_bundler" do
      Dir[File.join(slug_vendor_base, "**", ".git")].each do |dir|
        FileUtils.rm_rf(dir)
      end
      bundler.clean
    end
  end

  # writes ERB based database.yml for Rails. The database.yml uses the DATABASE_URL from the environment during runtime.
  def create_database_yml
    instrument 'ruby.create_database_yml' do
      return false unless File.directory?("config")
      return false if  bundler.has_gem?('activerecord') && bundler.gem_version('activerecord') >= Gem::Version.new('4.1.0.beta1')

      log("create_database_yml") do
        topic("Writing config/database.yml to read from DATABASE_URL")
        File.open("config/database.yml", "w") do |file|
          file.puts <<-DATABASE_YML
<%

require 'cgi'
require 'uri'

begin
  uri = URI.parse(ENV["DATABASE_URL"])
rescue URI::InvalidURIError
  raise "Invalid DATABASE_URL"
end

raise "No RACK_ENV or RAILS_ENV found" unless ENV["RAILS_ENV"] || ENV["RACK_ENV"]

def attribute(name, value, force_string = false)
  if value
    value_string =
      if force_string
        '"' + value + '"'
      else
        value
      end
    "\#{name}: \#{value_string}"
  else
    ""
  end
end

adapter = uri.scheme
adapter = "postgresql" if adapter == "postgres"

database = (uri.path || "").split("/")[1]

username = uri.user
password = uri.password

host = uri.host
port = uri.port

params = CGI.parse(uri.query || "")

%>

<%= ENV["RAILS_ENV"] || ENV["RACK_ENV"] %>:
  <%= attribute "adapter",  adapter %>
  <%= attribute "database", database %>
  <%= attribute "username", username %>
  <%= attribute "password", password, true %>
  <%= attribute "host",     host %>
  <%= attribute "port",     port %>

<% params.each do |key, value| %>
  <%= key %>: <%= value.first %>
<% end %>
          DATABASE_YML
        end
      end
    end
  end

  def rake
    @rake ||= begin
      raise_on_fail      = bundler.gem_version('railties') && bundler.gem_version('railties') > Gem::Version.new('3.x')

      topic "Detecting rake tasks"
      rake = LanguagePack::Helpers::RakeRunner.new()
      rake.load_rake_tasks!({ env: rake_env }, raise_on_fail)
      rake
    end
  end

  def rake_env
    if database_url
      { "DATABASE_URL" => database_url }
    else
      {}
    end.merge(user_env_hash)
  end

  def database_url
    env("DATABASE_URL") if env("DATABASE_URL")
  end

  # executes the block with GIT_DIR environment variable removed since it can mess with the current working directory git thinks it's in
  # @param [block] block to be executed in the GIT_DIR free context
  def allow_git(&blk)
    git_dir = ENV.delete("GIT_DIR") # can mess with bundler
    blk.call
    ENV["GIT_DIR"] = git_dir
  end

  # decides if we need to install the node.js binary
  # @note execjs will blow up if no JS RUNTIME is detected and is loaded.
  # @return [Array] the node.js binary path if we need it or an empty Array
  def add_node_js_binary
    bundler.has_gem?('execjs') && node_not_preinstalled? ? [@node_installer.binary_path] : []
  end

  # checks if node.js is installed
  # @return String if it's detected and false if it isn't
  def node_preinstall_bin_path
    return @node_preinstall_bin_path if defined?(@node_preinstall_bin_path)

    legacy_path = "#{Dir.pwd}/#{NODE_BP_PATH}"
    path        = run("which node")
    node_supplied = false
    if path && $?.success?
      node_supplied = true
      @node_preinstall_bin_path = path
    elsif run("#{legacy_path}/node -v") && $?.success?
      node_supplied = false
      @node_preinstall_bin_path = legacy_path
    else
      @node_preinstall_bin_path = false
    end
    if node_supplied
      puts "Skipping install of nodejs since it has been supplied"
    end
      @node_preinstall_bin_path
  end
  alias :node_js_installed? :node_preinstall_bin_path

  def node_not_preinstalled?
    !node_js_installed?
  end

  def run_assets_precompile_rake_task
    instrument 'ruby.run_assets_precompile_rake_task' do
      precompile = rake.task("assets:precompile")
      return true unless precompile.is_defined?

      topic "Precompiling assets"
      precompile.invoke(env: rake_env)
      if precompile.success?
        puts "Asset precompilation completed (#{"%.2f" % precompile.time}s)"
      else
        precompile_fail(precompile.output)
      end
    end
  end

  def precompile_fail(output)
    log "assets_precompile", :status => "failure"
    msg = "Precompiling assets failed.\n"
    if output.match(/(127\.0\.0\.1)|(org\.postgresql\.util)/)
      msg << "Attempted to access a nonexistent database:\n"
      msg << "https://docs.cloudfoundry.org/buildpacks/ruby/ruby-service-bindings.html\n"
    end
    error msg
  end

  def bundler_cache
    "vendor/bundle"
  end

  def load_bundler_cache
    instrument "ruby.load_bundler_cache" do
      cache.load "vendor"

      full_ruby_version       = run_stdout(%q(ruby -v)).chomp
      rubygems_version        = run_stdout(%q(gem -v)).chomp
      cf_metadata         = "vendor/.cloudfoundry/metadata"
      old_rubygems_version    = nil
      ruby_version_cache      = "ruby_version"
      buildpack_version_cache = "buildpack_version"
      cf_buildpack_version_cache = "cf_buildpack_version"
      bundler_version_cache   = "bundler_version"
      rubygems_version_cache  = "rubygems_version"
      stack_cache             = "stack"

      old_rubygems_version = @metadata.read(ruby_version_cache).chomp if @metadata.exists?(ruby_version_cache)
      old_stack = @metadata.read(stack_cache).chomp if @metadata.exists?(stack_cache)
      old_stack ||= "Unknown"

      stack_change  = old_stack != @stack
      convert_stack = @bundler_cache.old?
      @bundler_cache.convert_stack(stack_change) if convert_stack
      if !new_app? && stack_change
        puts "Purging Cache. Changing stack from #{old_stack} to #{@stack}"
        purge_bundler_cache(old_stack)
      elsif !new_app? && !convert_stack
        @bundler_cache.load
      end

      # fix bug from v37 deploy
      if File.exists?("vendor/ruby_version")
        puts "Broken cache detected. Purging build cache."
        cache.clear("vendor")
        FileUtils.rm_rf("vendor/ruby_version")
        purge_bundler_cache
        # fix bug introduced in v38
      elsif !@metadata.exists?(buildpack_version_cache) && @metadata.exists?(ruby_version_cache)
        puts "Broken cache detected. Purging build cache."
        purge_bundler_cache
      elsif (@bundler_cache.exists? || @bundler_cache.old?) && @metadata.exists?(ruby_version_cache) && full_ruby_version != @metadata.read(ruby_version_cache).chomp
        puts "Ruby version change detected. Clearing bundler cache."
        puts "Old: #{@metadata.read(ruby_version_cache).chomp}"
        puts "New: #{full_ruby_version}"
        purge_bundler_cache
      end

      # fix git gemspec bug from Bundler 1.3.0+ upgrade
      if File.exists?(bundler_cache) && !@metadata.exists?(bundler_version_cache) && !run("find vendor/bundle/*/*/bundler/gems/*/ -name *.gemspec").include?("No such file or directory")
        puts "Old bundler cache detected. Clearing bundler cache."
        purge_bundler_cache
      end

      # fix for https://github.com/sparklemotion/nokogiri/issues/923
      if @metadata.exists?(buildpack_version_cache) && (bv = @metadata.read(buildpack_version_cache).sub('v', '').to_i) && bv != 0 && bv <= 76
        puts "Fixing nokogiri install. Clearing bundler cache."
        puts "See https://github.com/sparklemotion/nokogiri/issues/923."
        purge_bundler_cache
      end

      FileUtils.mkdir_p(cf_metadata)
      @metadata.write(ruby_version_cache, full_ruby_version, false)
      @metadata.write(buildpack_version_cache, BUILDPACK_VERSION, false)
      @metadata.write(cf_buildpack_version_cache, CF_BUILDPACK_VERSION, false)
      @metadata.write(bundler_version_cache, BUNDLER_VERSION, false)
      @metadata.write(rubygems_version_cache, rubygems_version, false)
      @metadata.write(stack_cache, @stack, false)
      @metadata.save
    end
  end

  def purge_bundler_cache(stack = nil)
    instrument "ruby.purge_bundler_cache" do
      @bundler_cache.clear(stack)
      # need to reinstall language pack gems
      install_bundler_in_app
    end
  end
end
