package versioning

import (
	_ "embed"
)

//go:embed resources/sha
var sha string

//go:embed resources/build_time
var buildTime string

func GetBuildSha() string {
	return sha
}

func GetBuildTime() string {
	return buildTime
}
