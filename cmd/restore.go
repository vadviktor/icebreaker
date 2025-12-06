package cmd

import (
	"os"
	"time"

	"github.com/vadviktor/icebreaker/helpers"
	"github.com/vadviktor/icebreaker/restore"
	"github.com/vadviktor/icebreaker/types"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var logger = log.NewWithOptions(os.Stdout, log.Options{
	TimeFormat:      time.DateTime,
	ReportTimestamp: true,
})

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Initiates restoration of objects in S3 Glacier Deep Archive",
	Run: func(cmd *cobra.Command, args []string) {
		if err := helpers.ValidateFlags(cmd); err != nil {
			cmd.Help()
			os.Exit(0)
		}

		days, err := cmd.Flags().GetInt("days")
		if err != nil {
			days = 1
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			dryRun = false
		}

		var s3Parts types.S3Parts
		s3Parts, err = helpers.S3PartsFromUri(cmd.Flag("uri").Value.String())
		if err != nil {
			s3Parts = types.S3Parts{
				Bucket: cmd.Flag("bucket").Value.String(),
				Prefix: cmd.Flag("path").Value.String(),
			}
		}

		err = restore.RestoreObjects(&restore.RestoreConfig{
			Bucket: s3Parts.Bucket,
			Prefix: s3Parts.Prefix,
			Days:   days,
			DryRun: dryRun,
			Logger: logger,
		})
		if err != nil {
			logger.Error("Failed to restore objects: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	restoreCmd.Flags().StringP("uri", "u", "", "The S3 URI of the object or path (of objects) to restore. It can be a single object or a prefix of objects. E.g.: s3://mybucket/myfolder")
	restoreCmd.Flags().StringP("bucket", "b", "", "The S3 bucket to restore the objects from.")
	restoreCmd.Flags().StringP("path", "p", "", "The S3 path (a.k.a. prefix) to restore the objects from. It can be a single object or a prefix of objects.")

	restoreCmd.MarkFlagsMutuallyExclusive("uri", "bucket")
	restoreCmd.MarkFlagsMutuallyExclusive("uri", "path")

	restoreCmd.Flags().IntP("days", "d", 1, "The number of days to restore the objects for.")
	restoreCmd.Flags().BoolP("dry-run", "n", false, "List affected objects without restoring.")
}
