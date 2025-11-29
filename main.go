package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"icebreaker/s3restore"

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

	err := s3restore.RestoreObjects(s3restore.RestoreConfig{
		Bucket: bucket,
		Prefix: prefix,
		Days:   appCfg.days,
		DryRun: appCfg.dryRun,
		Logger: logger,
	})
	if err != nil {
		logger.Fatalf("Failed to restore objects: %v", err)
	}
}
