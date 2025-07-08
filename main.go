package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	charm_log "github.com/charmbracelet/log"
)

var logger = charm_log.NewWithOptions(os.Stdout, charm_log.Options{
	TimeFormat:      time.DateTime,
	ReportTimestamp: true,
})

const usageTemplate = `Usage of {{.ProgramName}}:

Restore objects from S3 Glacier Deep Archive

It iterates through objects at the specified S3 path, identifies objects in
Deep Archive, and initiates a restoration request for them if they are not
already restored or in the process of being restored.

Options:
{{.Flags}}
Examples:
  {{.ProgramName}} -path s3://mybucket/myfolder
  {{.ProgramName}} -path s3://mybucket/myfolder -days 7 -dry-run
`

func appUsage() {
	tmpl, err := template.New("usage").Parse(usageTemplate)
	if err != nil {
		logger.Fatal("Error parsing usage template:", err)
	}

	// Capture flag defaults output
	var flagsOutput strings.Builder

	flag.CommandLine.SetOutput(&flagsOutput)
	flag.PrintDefaults()
	flag.CommandLine.SetOutput(os.Stderr)

	data := struct {
		ProgramName string
		Flags       string
	}{
		ProgramName: filepath.Base(os.Args[0]),
		Flags:       flagsOutput.String(),
	}

	err = tmpl.Execute(os.Stderr, data)
	if err != nil {
		logger.Fatal("Error executing usage template:", err)
	}
}

type AppConfig struct {
	s3Path string
	days   int
	dryRun bool
}

func parseFlags() AppConfig {
	s3Path := flag.String("path", "", "The S3 path to restore (e.g. s3://mybucket/myfolder)")
	days := flag.Int("days", 1, "Number of days to restore objects for")
	dryRun := flag.Bool("dry-run", false, "List affected objects without restoring")

	flag.Usage = appUsage

	flag.Parse()

	if *s3Path == "" {
		logger.Error("Error: -path is required")
		flag.Usage()
		os.Exit(1)
	}

	if !strings.HasPrefix(*s3Path, "s3://") {
		logger.Error("Error: -path must start with s3://")
		flag.Usage()
		os.Exit(1)
	}

	return AppConfig{
		s3Path: *s3Path,
		days:   *days,
		dryRun: *dryRun,
	}
}

func main() {
	appCfg := parseFlags()

	pathParts := strings.SplitN(strings.TrimPrefix(appCfg.s3Path, "s3://"), "/", 2)
	bucket := pathParts[0]

	prefix := ""
	if len(pathParts) > 1 {
		prefix = pathParts[1]
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Fatalf("Unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
		OptionalObjectAttributes: []types.OptionalObjectAttributes{
			types.OptionalObjectAttributesRestoreStatus,
		},
	})

	logger.Infof("Processing objects in s3://%s/%s", bucket, prefix)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			logger.Fatalf("Failed to get page, %v", err)
		}

		for _, obj := range page.Contents {
			processObject(obj, s3Client, bucket, appCfg.days, appCfg.dryRun)
		}
	}

	logger.Info("Processing complete.")
}

func processObject(obj types.Object, s3Client *s3.Client, bucket string, days int, dryRun bool) {
	if obj.Key == nil {
		return
	}

	objectKey := *obj.Key

	if obj.StorageClass != types.ObjectStorageClass(types.StorageClassDeepArchive) {
		return
	}

	if dryRun {
		logger.Infof("🔍 Would restore: %s", objectKey)

		return
	}

	restoreStatus := obj.RestoreStatus

	if objectNotBeingRestored(restoreStatus) {
		logger.Infof("🚀 Requesting restoration: %s", objectKey)

		_, err := s3Client.RestoreObject(context.TODO(), &s3.RestoreObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
			RestoreRequest: &types.RestoreRequest{
				Days: aws.Int32(int32(days)),
				GlacierJobParameters: &types.GlacierJobParameters{
					Tier: types.TierBulk,
				},
			},
		})
		if err != nil {
			logger.Warnf("Failed to restore %s: %v", objectKey, err)
		}
	} else if objectIsRestored(restoreStatus) {
		expiryDate := "N/A"
		if restoreStatus.RestoreExpiryDate != nil {
			expiryDate = restoreStatus.RestoreExpiryDate.Format(time.RFC3339)
		}

		logger.Infof("✅ Restored: %s, ⌛ until: %s", objectKey, expiryDate)
	} else if restoreStatus != nil && restoreStatus.IsRestoreInProgress != nil && *restoreStatus.IsRestoreInProgress {
		logger.Infof("🏗️ Restoring: %s", objectKey)
	}
}

func objectNotBeingRestored(status *types.RestoreStatus) bool {
	if status == nil {
		return true // No status means not restored and not in progress
	}

	isRestoreInProgress := false
	if status.IsRestoreInProgress != nil {
		isRestoreInProgress = *status.IsRestoreInProgress
	}

	return !isRestoreInProgress && status.RestoreExpiryDate == nil
}

func objectIsRestored(status *types.RestoreStatus) bool {
	if status == nil {
		return false
	}

	isRestoreInProgress := false
	if status.IsRestoreInProgress != nil {
		isRestoreInProgress = *status.IsRestoreInProgress
	}

	return !isRestoreInProgress && status.RestoreExpiryDate != nil
}
