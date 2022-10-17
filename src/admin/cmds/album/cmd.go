package album

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "album",
		Short: "album operations",
	}
	cmd.AddCommand(
		HomePageCmd(),
		SearchPageCmd(),
		DownloadCmd(),
	)
	return cmd
}
