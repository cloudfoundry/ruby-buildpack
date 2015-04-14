require 'cf_spec_helper'

describe 'For all supported Ruby versions', if: ENV['CF_CI_ENV'] == "true" do
  shared_examples 'a Sinatra app' do
    let(:app) { Machete.deploy_app("rubies/tmp/#{ruby_version}/sinatra") }
    let(:browser) { Machete::Browser.new(app) }

    specify do
      generate_app('sinatra', ruby_version, engine, engine_version)
      assert_ruby_version_installed(ruby_version)

      unless engine == 'ruby'
        assert_ruby_version_and_engine_installed(ruby_version, engine, engine_version)
      end

      assert_root_contains('Hello, World')
      assert_offline_mode_has_no_traffic
    end
  end

  shared_examples 'a Rails 4 app' do
    let(:app) { Machete.deploy_app("rubies/tmp/#{ruby_version}/rails4", with_pg: true) }
    let(:browser) { Machete::Browser.new(app) }

    specify do
      generate_app('rails4', ruby_version, engine, engine_version)
      assert_ruby_version_installed(ruby_version)

      unless engine == 'ruby'
        assert_ruby_version_and_engine_installed(ruby_version, engine, engine_version)
      end

      assert_root_contains('The Kessel Run')
      assert_offline_mode_has_no_traffic
    end
  end

  context 'Ruby 1.9.3' do
    let(:ruby_version) { '1.9.3' }
    let(:engine) { 'ruby' }
    let(:engine_version) { ruby_version }

    it_behaves_like 'a Sinatra app'
    it_behaves_like 'a Rails 4 app'
  end

  context 'Ruby 2.0.0' do
    let(:ruby_version) { '2.0.0' }
    let(:engine) { 'ruby' }
    let(:engine_version) { ruby_version }

    it_behaves_like 'a Sinatra app'
    it_behaves_like 'a Rails 4 app'
  end

  context 'Ruby 2.1.5' do
    let(:ruby_version) { '2.1.5' }
    let(:engine) { 'ruby' }
    let(:engine_version) { ruby_version }

    it_behaves_like 'a Sinatra app'
    it_behaves_like 'a Rails 4 app'
  end

  context 'Ruby 2.2.0' do
    let(:ruby_version) { '2.2.0' }
    let(:engine) { 'ruby' }
    let(:engine_version) { ruby_version }

    it_behaves_like 'a Sinatra app'
    it_behaves_like 'a Rails 4 app'
  end

  context 'JRuby 1.7.11 Ruby 2.0.0' do
    let(:ruby_version) { '2.0.0' }
    let(:engine) { 'jruby' }
    let(:engine_version) { '1.7.11' }

    it_behaves_like 'a Sinatra app'
    it_behaves_like 'a Rails 4 app'
  end

  context 'JRuby 1.7.11 Ruby 1.9.3' do
    let(:ruby_version) { '1.9.3' }
    let(:engine) { 'jruby' }
    let(:engine_version) { '1.7.11' }

    it_behaves_like 'a Sinatra app'
    it_behaves_like 'a Rails 4 app'
  end

  context 'JRuby 1.7.11 Ruby 1.8.7' do
    let(:ruby_version) { '1.8.7' }
    let(:engine) { 'jruby' }
    let(:engine_version) { '1.7.11' }

    it_behaves_like 'a Sinatra app'
    # Rails app not tested here - requires a minimum of Ruby 1.9.3
  end

  def evaluate_erb(file_path, our_binding)
    template = File.read(file_path)
    f = File.open(file_path, 'w')
    f << ERB.new(template).result(our_binding)
    f.close
  end

  def assert_offline_mode_has_no_traffic
    expect(app.host).not_to have_internet_traffic if Machete::BuildpackMode.offline?
  end

  def generate_app(app_name, ruby_version, engine, engine_version)
    origin_template_path = File.join(File.dirname(__FILE__), '..', 'fixtures', 'rubies', app_name)
    copied_template_path = File.join(File.dirname(__FILE__), '..', 'fixtures', 'rubies', 'tmp', ruby_version, app_name)
    FileUtils.rm_rf(copied_template_path)
    FileUtils.mkdir_p(File.dirname(copied_template_path))
    FileUtils.cp_r(origin_template_path, copied_template_path)

    ['Gemfile', 'package.sh', '.jrubyrc'].each do |file|
      evaluate_erb(File.join(copied_template_path, file), binding)
    end
  end

  def assert_ruby_version_installed(ruby_version)
    expect(app).to be_running
    expect(app).to have_logged "Using Ruby version: ruby-#{ruby_version}"
  end

  def assert_ruby_version_and_engine_installed(ruby_version, engine, engine_version)
    expect(app).to be_running
    expect(app).to have_logged "Using Ruby version: ruby-#{ruby_version}-#{engine}-#{engine_version}"
  end

  def assert_root_contains(text)
    browser.visit_path('/')
    expect(browser).to have_body(text)
  end
end
