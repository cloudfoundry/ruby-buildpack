require 'cf_spec_helper'

describe 'When running ./bin/compile' do
  def run(cmd, env: {})
    if RUBY_PLATFORM =~ /darwin/i
      env_flags = env.map{|k,v| "-e #{k}=#{v}"}.join(' ')
      `docker run --rm #{env_flags} -v #{Dir.pwd}:/buildpack:ro -w /buildpack cloudfoundry/cflinuxfs2 #{cmd}`
    else
      `env #{env.map{|k,v| "#{k}=#{v}"}.join(' ')} #{cmd}`
    end
  end

  context 'and on an unsupported stack' do
    before(:all) do
      @output = run("./bin/compile #{Dir.mktmpdir} #{Dir.mktmpdir} 2>&1", env: {CF_STACK: 'unsupported'})
    end

    it 'displays a helpful error message' do
      expect(@output).to include('not supported by this buildpack')
    end

    it 'exits with our error code' do
      expect($?.exitstatus).to eq 44
    end
  end

  context 'and on a supported stack' do
    before(:all) do
      @output = run("./bin/compile #{Dir.mktmpdir} #{Dir.mktmpdir} 2>&1", env: {CF_STACK: 'cflinuxfs2'})
    end

    it 'does not display an error message' do
      expect(@output).to_not include('not supported by this buildpack')
    end

    it 'does not exit with our error code' do
      expect($?.exitstatus).to_not eq 44
    end
  end
end
