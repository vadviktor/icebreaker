package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

func ValidateFlags(cmd *cobra.Command) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	uri, err := cmd.Flags().GetString("uri")
	if err != nil {
		return err
	}
	err = validate.Var(uri, "required,uri,startswith=s3://")
	if err != nil {
		return err
	}

	// bucket, err := cmd.Flags().GetString("bucket")
	// if err != nil {
	// 	return err
	// }

	// path, err := cmd.Flags().GetString("path")
	// if err != nil {
	// 	return err
	// }

	// if uri == "" && (bucket == "" || path == "") {
	// 	return fmt.Errorf("[uri] or [bucket] and [path] are required")
	// }

	// if uri != "" && !strings.HasPrefix(uri, "s3://") {
	// 	return fmt.Errorf("[uri] must start with s3://")
	// }

	return nil
}
