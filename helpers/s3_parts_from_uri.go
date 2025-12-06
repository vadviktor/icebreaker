package helpers

import (
	"fmt"
	"strings"

	"github.com/vadviktor/icebreaker/types"
)

func S3PartsFromUri(uri string) (types.S3Parts, error) {
	if uri == "" || !strings.HasPrefix(uri, "s3://") {
		return types.S3Parts{}, fmt.Errorf("invalid URI: %s", uri)
	}

	pathParts := strings.SplitN(strings.TrimPrefix(uri, "s3://"), "/", 2)
	bucket := pathParts[0]

	prefix := ""
	if len(pathParts) > 1 {
		prefix = pathParts[1]
	}

	return types.S3Parts{
		Bucket: bucket,
		Prefix: prefix,
	}, nil
}
