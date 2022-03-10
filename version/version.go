package version

var (
	buildTime string
	commit    string
	release   string
)

type Version struct {
	BuildTime string
	Commit    string
	Release   string
}

func GetVersion() Version {
	return *&Version{
		BuildTime: buildTime,
		Commit:    commit,
		Release:   release,
	}
}
