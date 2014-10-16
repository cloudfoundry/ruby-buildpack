require_relative '../cf_spec_helper'

describe "DependencyExistenceChecker" do
  describe ".exists?" do
    subject { DependencyExistenceChecker.exists?(original_filename) }

    before do
      stub_const("DEPENDENCIES_PATH", "some/dependencies/path")
      allow(File).to receive(:exists?) do |arg|
        arg == 'some/dependencies/path/file.zip'
      end
    end

    context "when the given file is at the dependency path" do
      let(:original_filename) { "file.zip" }
      it {is_expected.to eq true}
    end

    context "when the given file is not at the dependency path" do
      let(:original_filename) { "nonexistent_file.zip" }
      it {is_expected.to eq false}
    end
  end
end

