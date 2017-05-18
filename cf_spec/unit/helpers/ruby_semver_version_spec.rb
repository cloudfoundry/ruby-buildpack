require 'cf_spec_helper'
require 'language_pack/ruby_semver_version'

describe LanguagePack::RubySemverVersion do
  let(:fixtures) { File.join(File.dirname(__FILE__), '..', '..', 'fixtures', 'manifests') }

  describe "#version" do
    subject { described_class.new(gemfile, manifest_file).version }

    context "simple manifest" do
      let(:manifest_file) { File.join(fixtures, 'simple_manifest.yml') }

      context "~> gemfile" do
        let(:gemfile) { File.join(fixtures, 'Gemfile_8_2_0') }
        it { should eq('8.2.7') }
      end

      context "exact and supplied gemfile" do
        let(:gemfile) { File.join(fixtures, 'Gemfile_8_2_7') }
        it { should eq('8.2.7') }
      end

      context "exact and NOT supplied gemfile" do
        let(:gemfile) { File.join(fixtures, 'Gemfile_8_2_9') }
        it { should eq('') }
      end

      context "gemfile with global ruby constants" do
        let(:gemfile) { File.join(fixtures, 'Gemfile_with_const') }
        it "don't raise an error" do
          expect { subject }.to_not raise_error
        end
      end

      context "gemfile with global ruby constants in conditionals" do
        let(:gemfile) { File.join(fixtures, 'Gemfile_with_constants_in_conditionals') }
        it "don't raise an error" do
          expect { subject }.to_not raise_error
        end
        it { should eq('8.2.7') }
      end
    end

    context "unordered manifest" do
      let(:manifest_file) { File.join(fixtures, 'unordered_manifest.yml') }

      context "~> gemfile" do
        let(:gemfile) { File.join(fixtures, 'Gemfile_8_2_0') }
        it { should eq('8.2.22') }
      end

      context "~> gemfile (major/minor only)" do
        let(:gemfile) { File.join(fixtures, 'Gemfile_8_2') }
        it { should eq('8.6.9') }
      end

      context "complicated ruby line gemfile" do
        let(:gemfile) { File.join(fixtures, 'Gemfile_complicated_ruby') }
        it { should eq('8.2.22') }
      end
    end

    context "complicated manifest" do
      let(:manifest_file) { File.join(fixtures, 'complicated_manifest.yml') }

      context "specified_ruby_version fixture" do
        let(:gemfile) { File.join(fixtures, '..', 'specified_ruby_version', 'Gemfile') }
        it { should eq('2.2.6') }
      end

      context "unspecified_ruby fixture" do
        let(:gemfile) { File.join(fixtures, '..', 'unspecified_ruby', 'Gemfile') }
        it 'finds a ruby matching "~> DEFAULT_VERSION_NUMBER"' do
          expect(subject).to eq('2.4.1')
        end
      end
    end
  end
end
