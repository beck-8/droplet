package version

var (
	CurrentCommit string

	Version = "v2.12.0"
)

func UserVersion() string {
	return Version + CurrentCommit
}
