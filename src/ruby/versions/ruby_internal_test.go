package versions

import "testing"

// TestParseBundlerVersion is a regression test for a real CI failure caused
// by a rubygems dependency bump (3.4.19 -> 4.0.22): `bundle version`'s
// output format changed, dropping the "Bundler version " prefix, which broke
// the previous regex entirely (zero matches, hard failure during staging).
func TestParseBundlerVersion(t *testing.T) {
	cases := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name:   "legacy format with 'Bundler version ' prefix",
			output: "Bundler version 2.7.2 (2025-09-09 commit b463ced1459)\n",
			want:   "2.7.2",
		},
		{
			name:   "new format without prefix (RubyGems/Bundler 4.x)",
			output: "4.0.22 (2026-09-30 commit ff2bd50)\n",
			want:   "4.0.22",
		},
		{
			name:   "bare version number with no trailing metadata",
			output: "1.17.2\n",
			want:   "1.17.2",
		},
		{
			name:    "unrecognized output",
			output:  "error: something went wrong\n",
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseBundlerVersion(tc.output)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got version %q", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
