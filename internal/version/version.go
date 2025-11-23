package version

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

var CompatibleClientVersions = []string{
	"1.0.0",
	"dev",
}

func IsCompatible(clientVersion string) bool {
	for _, v := range CompatibleClientVersions {
		if v == clientVersion {
			return true
		}
	}
	return false
}
