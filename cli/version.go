package cli

import (
	"edgeproxy/version"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show Application Version",
		Long:  `Show Application Version`,
		Run: func(cmd *cobra.Command, testSuites []string) {
			ver := version.GetVersion()

			b, _ := json.Marshal(ver)
			fmt.Printf(string(b))
			os.Exit(exitCode)
		},
	}
)

func init() {
	RootCmd.AddCommand(versionCmd)
}
