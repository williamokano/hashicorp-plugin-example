package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			name:  "valid semantic version",
			input: "1.2.3",
			want:  &Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "version with zeros",
			input: "0.0.1",
			want:  &Version{Major: 0, Minor: 0, Patch: 1},
		},
		{
			name:  "large version numbers",
			input: "100.200.300",
			want:  &Version{Major: 100, Minor: 200, Patch: 300},
		},
		{
			name:    "invalid format - too few parts",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "invalid format - too many parts",
			input:   "1.2.3.4",
			wantErr: true,
		},
		{
			name:    "invalid major version",
			input:   "a.2.3",
			wantErr: true,
		},
		{
			name:    "invalid minor version",
			input:   "1.b.3",
			wantErr: true,
		},
		{
			name:    "invalid patch version",
			input:   "1.2.c",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "version with v prefix",
			input:   "v1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version *Version
		want    string
	}{
		{
			name:    "standard version",
			version: &Version{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3",
		},
		{
			name:    "version with zeros",
			version: &Version{Major: 0, Minor: 0, Patch: 0},
			want:    "0.0.0",
		},
		{
			name:    "large numbers",
			version: &Version{Major: 100, Minor: 200, Patch: 300},
			want:    "100.200.300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		name string
		v1   *Version
		v2   *Version
		want int
		desc string
	}{
		{
			name: "equal versions",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3},
			want: 0,
			desc: "should return 0",
		},
		{
			name: "v1 major version higher",
			v1:   &Version{Major: 2, Minor: 0, Patch: 0},
			v2:   &Version{Major: 1, Minor: 9, Patch: 9},
			want: 1,
			desc: "should return positive",
		},
		{
			name: "v1 major version lower",
			v1:   &Version{Major: 1, Minor: 9, Patch: 9},
			v2:   &Version{Major: 2, Minor: 0, Patch: 0},
			want: -1,
			desc: "should return negative",
		},
		{
			name: "same major, v1 minor higher",
			v1:   &Version{Major: 1, Minor: 3, Patch: 0},
			v2:   &Version{Major: 1, Minor: 2, Patch: 9},
			want: 1,
			desc: "should return positive",
		},
		{
			name: "same major, v1 minor lower",
			v1:   &Version{Major: 1, Minor: 2, Patch: 9},
			v2:   &Version{Major: 1, Minor: 3, Patch: 0},
			want: -1,
			desc: "should return negative",
		},
		{
			name: "same major and minor, v1 patch higher",
			v1:   &Version{Major: 1, Minor: 2, Patch: 4},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3},
			want: 1,
			desc: "should return positive",
		},
		{
			name: "same major and minor, v1 patch lower",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3},
			v2:   &Version{Major: 1, Minor: 2, Patch: 4},
			want: -1,
			desc: "should return negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Compare(tt.v2)

			// We care about the sign, not the exact value
			if tt.want == 0 {
				assert.Equal(t, 0, got, tt.desc)
			} else if tt.want > 0 {
				assert.Greater(t, got, 0, tt.desc)
			} else {
				assert.Less(t, got, 0, tt.desc)
			}
		})
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name       string
		cliVersion string
		minVersion string
		maxVersion string
		want       bool
		wantErr    bool
	}{
		{
			name:       "compatible - within range",
			cliVersion: "1.5.0",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			want:       true,
		},
		{
			name:       "compatible - at minimum",
			cliVersion: "1.0.0",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			want:       true,
		},
		{
			name:       "compatible - at maximum",
			cliVersion: "2.0.0",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			want:       true,
		},
		{
			name:       "incompatible - below minimum",
			cliVersion: "0.9.0",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			want:       false,
		},
		{
			name:       "incompatible - above maximum",
			cliVersion: "2.1.0",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			want:       false,
		},
		{
			name:       "compatible - no minimum",
			cliVersion: "0.1.0",
			minVersion: "",
			maxVersion: "2.0.0",
			want:       true,
		},
		{
			name:       "compatible - no maximum",
			cliVersion: "10.0.0",
			minVersion: "1.0.0",
			maxVersion: "",
			want:       true,
		},
		{
			name:       "compatible - no constraints",
			cliVersion: "1.0.0",
			minVersion: "",
			maxVersion: "",
			want:       true,
		},
		{
			name:       "invalid CLI version",
			cliVersion: "invalid",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			wantErr:    true,
		},
		{
			name:       "invalid minimum version",
			cliVersion: "1.5.0",
			minVersion: "invalid",
			maxVersion: "2.0.0",
			wantErr:    true,
		},
		{
			name:       "invalid maximum version",
			cliVersion: "1.5.0",
			minVersion: "1.0.0",
			maxVersion: "invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsCompatible(tt.cliVersion, tt.minVersion, tt.maxVersion)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersionConstants(t *testing.T) {
	// Ensure version constants are set properly
	assert.NotEmpty(t, CLIVersion, "CLIVersion should not be empty")
	assert.Regexp(t, `^\d+\.\d+\.\d+$`, CLIVersion, "CLIVersion should be semantic version")

	// Test that CLI version is parseable
	v, err := Parse(CLIVersion)
	require.NoError(t, err, "CLIVersion should be parseable")
	assert.NotNil(t, v)
}
