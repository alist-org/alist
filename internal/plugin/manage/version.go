package plugin_manage

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// api版本规则
// 大版本.增强版本.修补版本-描述
//
// 增强版本在原基础上增加接口时增长
// 修补版本在依赖变化时增长
var PLUGIN_API_VERSION = ParseVersion("v1.0.0-test")

type Version struct {
	Major int
	Minor int
	Patch int
	Desc  string
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Desc)
}

var vreg = regexp.MustCompile(`v(\d+)\.?(\d+)\.?(\d*)[-_]?(\w*)`)

func ParseVersion(version string) Version {
	v := vreg.FindStringSubmatch(version)
	if len(v) == 0 {
		return Version{}
	}
	major, _ := strconv.Atoi(v[1])
	minor, _ := strconv.Atoi(v[2])
	patch, _ := strconv.Atoi(v[3])

	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Desc:  v[4],
	}
}

const (
	VersionSmall uint8 = iota
	VersionEqual
	VersionBig

	VersionIncompatible
	VersionCompatible
)

func CompareVersionStr(src, dest string) uint8 {
	return CompareVersion(ParseVersion(src), ParseVersion(dest))
}

// 比较 Major Minor Patch
// (src < dest) => VersionSmall
// (src == dest) => VersionEqual
// (src > dest) => VersionBig
func CompareVersion(src, dest Version) uint8 {
	if src.Major > dest.Major {
		return VersionBig
	} else if src.Major < dest.Major {
		return VersionSmall
	}

	if src.Minor > dest.Minor {
		return VersionBig
	} else if src.Minor < dest.Minor {
		return VersionSmall
	}

	if src.Patch > dest.Patch {
		return VersionBig
	} else if src.Patch < dest.Patch {
		return VersionSmall
	}
	return VersionEqual
}

// 比较 Major Minor 跳过 Patch
// (src < dest) => VersionSmall
// (src == dest) => VersionEqual
// (src > dest) => VersionBig
func CompareVersion2(src, dest Version) uint8 {
	if src.Major > dest.Major {
		return VersionBig
	} else if src.Major < dest.Major {
		return VersionSmall
	}

	if src.Minor > dest.Minor {
		return VersionBig
	} else if src.Minor < dest.Minor {
		return VersionSmall
	}
	return VersionEqual
}

func ComparePluginApiVersion(pluginV Version) uint8 {
	if pluginV.Major != PLUGIN_API_VERSION.Major {
		return VersionIncompatible
	}
	if pluginV.Minor > PLUGIN_API_VERSION.Minor {
		return VersionIncompatible
	}
	if pluginV.Patch > PLUGIN_API_VERSION.Patch {
		return VersionIncompatible
	}
	return VersionCompatible
}

func IsSupportPlugin(apiVersion string) bool {
	apiVersions := strings.Split(apiVersion, ",")
	for _, apiVersion := range apiVersions {
		if ComparePluginApiVersion(ParseVersion(apiVersion)) == VersionCompatible {
			return true
		}
	}
	return false
}
