package versioning

import "runtime/debug"

func GetBuildSha() string {
	sha := "unknown"

	buildInfo, ok := debug.ReadBuildInfo()
	if ok && buildInfo.Main.Sum != "" {
		sha = buildInfo.Main.Sum
	}

	return sha
}
