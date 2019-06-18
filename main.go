package main

import (
	"flag"
	"github.com/railsware/s3_fetch_certs/aws"
	"github.com/railsware/s3_fetch_certs/log"
	stdlog "log"
	"os"
	"os/exec"
	"strings"
)

var (
	accessKeyId, secretAccessKey, sessionToken, region, bucket, certsKey, outDirectory, outName, runAfterChange string
)

func main() {
	flag.StringVar(&accessKeyId, "accessKeyId", "", "AWS access key ID")
	flag.StringVar(&secretAccessKey, "secretAccessKey", "", "AWS secret access key")
	flag.StringVar(&sessionToken, "sessionToken", "", "AWS session token")
	flag.StringVar(&region, "region", "us-east-1", "AWS region")
	flag.StringVar(&bucket, "bucket", "", "AWS S3 bucket")
	flag.StringVar(&certsKey, "certsKey", "", "AWS certs key")
	flag.StringVar(&outDirectory, "outDirectory", "", "Output directory")
	flag.StringVar(&outName, "outName", "certificate", "Output name without extension")
	flag.StringVar(&runAfterChange, "runAfterChange", "", "Run command after change")

	flag.Parse()
	// logger
	log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
	// conf
	log.StartupInfo()
	// aws client
	awsClient, error := aws.InitAWSClient(accessKeyId, secretAccessKey, sessionToken, region)

	if error != nil {
		return
	}

	if awsClient.DownloadFiles(bucket, certsKey, outDirectory, outName) && runAfterChange != "" {
		command := strings.Split(runAfterChange, " ")
		cmd := exec.Command(command[0])
		if len(command) > 1 {
			cmd = exec.Command(command[0], command[1:]...)
		}

		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("cmd.CombinedOutput() failed with %s\n", err)
		}
		log.Infof("Command result: %s\n", stdoutStderr)
	}
}
