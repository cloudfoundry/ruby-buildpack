require 'cf_spec_helper'
require 'open3'

describe 'CF Ruby Buildpack' do
  before(:all) do
    @app = Machete.deploy_app('specified_ruby_version')
  end

  after(:all) { Machete::CF::DeleteApp.new.execute(@app) }

  it 'starts' do
    expect(@app).to be_running
  end

  it 'uses the specified ruby version' do
    expect(@app).to have_logged 'Using Ruby version: ruby-2.2.7'
  end

  context 'running a task' do
    before { skip_if_no_run_task_support_on_targeted_cf }

    it 'can find the specifed ruby in the container' do
      expect(@app).to be_running

      Open3.capture2e('cf','run-task', 'specified_ruby_version', 'echo "RUNNING A TASK: $(ruby --version)"')[1].success? or raise 'Could not create run task'
      expect(@app).to have_logged(/RUNNING A TASK: ruby 2\.2\.7/)
    end
  end
end
