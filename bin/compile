#!/usr/bin/env ruby

# sync output
$stdout.sync = true

cache_dir = ARGV[1]
compile_extensions_dir = File.expand_path(File.join(File.dirname(__FILE__), '..', 'compile-extensions'))

exit(44) unless system("#{compile_extensions_dir}/bin/check_stack_support")

$:.unshift File.expand_path("../../lib", __FILE__)
require "language_pack"
require "language_pack/shell_helpers"

require 'cloud_foundry/language_pack/extensions'

begin
  LanguagePack::Instrument.trace 'compile', 'app.compile' do
    if pack = LanguagePack.detect(ARGV[0], ARGV[1])
      LanguagePack::ShellHelpers.initialize_env(ARGV[2])
      pack.topic("Compiling #{pack.name}")
      pack.log("compile") do
        pack.compile
      end
    end
  end
rescue Exception => e
  Kernel.puts " !"
  e.message.split("\n").each do |line|
    Kernel.puts " !     #{line.strip}"
  end
  Kernel.puts " !"
  if e.is_a?(BuildpackError)
    exit 1
  else
    raise e
  end
end
