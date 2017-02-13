require 'cf_spec_helper'

describe 'When running detect script' do

  context 'project root directory has Gemfile' do
    before do
      root_dir = Dir.pwd
      Dir.chdir('./cf_spec/fixtures/rails4') do
        @output = `#{root_dir}/bin/detect $PWD 2>&1`
      end
    end

    it 'displays the current ruby buildpack version' do
      expect(@output).to include('ruby ' + File.read('VERSION'))
    end

    it 'exits with error code 0' do
      expect($?.exitstatus).to eq 0
    end
  end

  context 'project root directory does not have Gemfile' do
    before do
      root_dir = Dir.pwd
      Dir.chdir('./cf_spec/fixtures/no_gemfile') do
        @output = `#{root_dir}/bin/detect $PWD 2>&1`
      end
    end

    it 'does not recognize as ruby app' do
      expect(@output).to include('no')
    end

    it 'exits with error code 1' do
      expect($?.exitstatus).to eq 1
    end
  end
end
