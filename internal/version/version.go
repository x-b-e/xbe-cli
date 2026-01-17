package version

// Version is overridden at build time via -ldflags.
// Default is "dev".
var Version = "dev"

func String() string {
	return Version
}
