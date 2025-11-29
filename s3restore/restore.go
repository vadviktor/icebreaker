package s3restore

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	charm_log "github.com/charmbracelet/log"
)

// RestoreConfig holds the configuration for restoring S3 objects from Glacier Deep Archive
type RestoreConfig struct {
	Bucket string
	Prefix string
	Days   int
	DryRun bool
	Logger *charm_log.Logger
}

// RestoreObjects iterates through objects at the specified S3 path, identifies objects in
// Deep Archive, and initiates a restoration request for them if they are not
// already restored or in the process of being restored.
func RestoreObjects(cfg RestoreConfig) error {
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	s3Client := s3.NewFromConfig(awsCfg)

	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(cfg.Bucket),
		Prefix: aws.String(cfg.Prefix),
		OptionalObjectAttributes: []types.OptionalObjectAttributes{
			types.OptionalObjectAttributesRestoreStatus,
		},
	})

	cfg.Logger.Infof("Processing objects in s3://%s/%s", cfg.Bucket, cfg.Prefix)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}

		for _, obj := range page.Contents {
			if err := processObject(obj, s3Client, cfg); err != nil {
				return err
			}
		}
	}

	cfg.Logger.Info("Processing complete.")

	return nil
}

func processObject(obj types.Object, s3Client *s3.Client, cfg RestoreConfig) error {
	if obj.Key == nil {
		return nil
	}

	objectKey := *obj.Key

	if obj.StorageClass != types.ObjectStorageClass(types.StorageClassDeepArchive) {
		return nil
	}

	if cfg.DryRun {
		cfg.Logger.Infof("🔍 Would restore: %s", objectKey)
		return nil
	}

	restoreStatus := obj.RestoreStatus

	if objectNotBeingRestored(restoreStatus) {
		cfg.Logger.Infof("🚀 Requesting restoration: %s", objectKey)

		_, err := s3Client.RestoreObject(context.TODO(), &s3.RestoreObjectInput{
			Bucket: aws.String(cfg.Bucket),
			Key:    aws.String(objectKey),
			RestoreRequest: &types.RestoreRequest{
				Days: aws.Int32(int32(cfg.Days)),
				GlacierJobParameters: &types.GlacierJobParameters{
					Tier: types.TierBulk,
				},
			},
		})
		if err != nil {
			return err
		}
	} else if objectIsRestored(restoreStatus) {
		expiryDate := "N/A"
		if restoreStatus.RestoreExpiryDate != nil {
			expiryDate = restoreStatus.RestoreExpiryDate.Format(time.RFC3339)
		}

		cfg.Logger.Infof("✅ Restored: %s, ⌛ until: %s", objectKey, expiryDate)
	} else if restoreStatus != nil && restoreStatus.IsRestoreInProgress != nil && *restoreStatus.IsRestoreInProgress {
		cfg.Logger.Infof("🏗️ Restoring: %s", objectKey)
	}

	return nil
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
