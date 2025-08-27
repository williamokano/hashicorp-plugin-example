package version

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	CLIVersion   = "1.0.0"
	CLIBuildTime = "unknown"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func Parse(v string) (*Version, error) {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return &Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

func IsCompatible(cliVersion, minVersion, maxVersion string) (bool, error) {
	cli, err := Parse(cliVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse CLI version: %w", err)
	}

	if minVersion != "" {
		min, err := Parse(minVersion)
		if err != nil {
			return false, fmt.Errorf("failed to parse minimum version: %w", err)
		}
		if cli.Compare(min) < 0 {
			return false, nil
		}
	}

	if maxVersion != "" {
		max, err := Parse(maxVersion)
		if err != nil {
			return false, fmt.Errorf("failed to parse maximum version: %w", err)
		}
		if cli.Compare(max) > 0 {
			return false, nil
		}
	}

	return true, nil
}
