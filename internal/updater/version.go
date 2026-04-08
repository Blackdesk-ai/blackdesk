package updater

import (
	"strconv"
	"strings"
)

type versionParts struct {
	major int
	minor int
	patch int
	ok    bool
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	return version
}

func versionLabel(version string) string {
	version = normalizeVersion(version)
	if version == "" {
		return "unknown"
	}
	if version == "dev" {
		return "dev"
	}
	return "v" + version
}

func parseVersion(version string) versionParts {
	version = normalizeVersion(version)
	if version == "" || version == "dev" {
		return versionParts{}
	}
	version = strings.SplitN(version, "+", 2)[0]
	version = strings.SplitN(version, "-", 2)[0]
	segments := strings.Split(version, ".")
	if len(segments) == 0 {
		return versionParts{}
	}

	values := []int{0, 0, 0}
	for i := 0; i < len(values) && i < len(segments); i++ {
		value, err := strconv.Atoi(segments[i])
		if err != nil {
			return versionParts{}
		}
		values[i] = value
	}

	return versionParts{
		major: values[0],
		minor: values[1],
		patch: values[2],
		ok:    true,
	}
}

func compareVersions(left, right string) (int, bool) {
	a := parseVersion(left)
	b := parseVersion(right)
	if !a.ok || !b.ok {
		return 0, false
	}
	if a.major != b.major {
		if a.major < b.major {
			return -1, true
		}
		return 1, true
	}
	if a.minor != b.minor {
		if a.minor < b.minor {
			return -1, true
		}
		return 1, true
	}
	if a.patch != b.patch {
		if a.patch < b.patch {
			return -1, true
		}
		return 1, true
	}
	return 0, true
}
