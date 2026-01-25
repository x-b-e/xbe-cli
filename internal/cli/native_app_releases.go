package cli

import "github.com/spf13/cobra"

var nativeAppReleasesCmd = &cobra.Command{
	Use:   "native-app-releases",
	Short: "Browse and view native app releases",
	Long: `Browse and view native app releases.

Native app releases capture mobile build metadata, git references, and
release channel status for iOS and Android apps.

Commands:
  list    List native app releases with filtering and pagination
  show    View the full details of a specific native app release`,
	Example: `  # List native app releases
  xbe view native-app-releases list

  # Filter by release channel
  xbe view native-app-releases list --release-channel apple-app-store

  # View a native app release
  xbe view native-app-releases show 123`,
}

func init() {
	viewCmd.AddCommand(nativeAppReleasesCmd)
}
