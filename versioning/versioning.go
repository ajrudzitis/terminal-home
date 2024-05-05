package versioning

import "runtime/debug"

func GetBuildSha() string {
	sha := "unknown"

	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		sha = buildInfo.Main.Version
	}

	return sha
}
