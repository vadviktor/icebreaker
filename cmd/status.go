package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vadviktor/icebreaker/helpers"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Reports the status of objects in S3 Glacier Deep Archive",
	Run: func(cmd *cobra.Command, args []string) {
		if err := helpers.ValidateFlags(cmd); err != nil {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	statusCmd.Flags().StringP("uri", "u", "", "The S3 URI of the object or path (of objects) to get the status of. It can be a single object or a prefix of objects. E.g.: s3://mybucket/myfolder")
	statusCmd.Flags().StringP("bucket", "b", "", "The S3 bucket to get the status of the objects in.")
	statusCmd.Flags().StringP("path", "p", "", "The S3 path (a.k.a. prefix) to get the status of the objects in. It can be a single object or a prefix of objects.")

	statusCmd.MarkFlagsMutuallyExclusive("uri", "bucket")
	statusCmd.MarkFlagsMutuallyExclusive("uri", "path")
}
