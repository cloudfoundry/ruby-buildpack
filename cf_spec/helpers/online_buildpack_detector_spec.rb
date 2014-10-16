require_relative '../cf_spec_helper'

describe "OnlineBuildpackDetector" do
  describe ".online?" do
    subject { OnlineBuildpackDetector.online? }
    let(:dir) { class_double(Dir).as_stubbed_const }

    before do
      allow(dir).to receive(:exist?).with("some/dependencies/path").and_return(exist)
      stub_const("DEPENDENCIES_PATH", "some/dependencies/path")
    end

    context "when there is a directory with the dependency path" do
      let(:exist) { true }
      it {is_expected.to eq false}
    end

    context "when there is not a directory with the dependency path" do
      let(:exist) { false }
      it {is_expected.to eq true}
    end
  end
end
