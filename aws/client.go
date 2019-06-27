package aws

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/railsware/s3_fetch_certs/log"
	"io"
	"os"
)

type AWSConnection struct {
	Session  *session.Session
	S3Client *s3.S3
}

func InitAWSClient(accessKeyId, secretAccessKey, sessionToken, region string) (*AWSConnection, error) {
	awsCredentials := credentials.NewEnvCredentials()

	if accessKeyId != "" && secretAccessKey != "" {
		awsCredentials = credentials.NewStaticCredentials(accessKeyId, secretAccessKey, sessionToken)
	}

	s3Config := &aws.Config{
		Credentials: awsCredentials,
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		MaxRetries:  aws.Int(5),
	}

	awsSession, err := session.NewSession(s3Config)

	if err != nil {
		log.Errorf("Problem with connection to aws: %v", err)
		return nil, err
	}

	s3Client := s3.New(awsSession)

	return &AWSConnection{
		Session:  awsSession,
		S3Client: s3Client,
	}, nil
}

// upload message to S3
func (awsc *AWSConnection) DownloadFiles(bucket, certsKey, outDirectory, outName string) bool {
	downloader := s3manager.NewDownloader(awsc.Session)

	objData, err := awsc.S3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("%s.crt", certsKey)),
	})

	if err != nil {
		log.Errorf("Unable to find cert on s3 %v", err)
		return false
	}

	s3FileSha256 := *objData.Metadata["Sha256"]

	certFilename := fmt.Sprintf("%s/%s.crt", outDirectory, outName)

	if s3FileSha256 != "" {
		if _, err := os.Stat(certFilename); !os.IsNotExist(err) {
			crtShaSumCheckKey, err := os.Open(certFilename)
			if err != nil {
				log.Errorf("Unable to open file %v", err)
				return false
			}

			hash := sha256.New()
			if _, err := io.Copy(hash, crtShaSumCheckKey); err != nil {
				log.Errorf("Unable to calculate file sha256 %v", err)
				return false
			}
			//Convert the bytes to a string
			hashSum := hex.EncodeToString(hash.Sum(nil))

			if hashSum == s3FileSha256 {
				log.Infof("%s sha256 the same as on s3", certFilename)
				return false
			}
		}
	}

	privateKey, err := os.Create(fmt.Sprintf("%s/%s.key", outDirectory, outName))
	if err != nil {
		log.Errorf("Unable to open file %v", err)
		return false
	}
	defer privateKey.Close()

	crtKey, err := os.Create(certFilename)
	if err != nil {
		log.Errorf("Unable to open file %v", err)
		return false
	}
	defer crtKey.Close()

	objects := []s3manager.BatchDownloadObject{
		{
			Object: &s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(fmt.Sprintf("%s.key", certsKey)),
			},
			Writer: privateKey,
		},
		{
			Object: &s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(fmt.Sprintf("%s.crt", certsKey)),
			},
			Writer: crtKey,
		},
	}

	iter := &s3manager.DownloadObjectsIterator{Objects: objects}
	if err := downloader.DownloadWithIterator(aws.BackgroundContext(), iter); err != nil {
		log.Errorf("Unable to download files %v", err)
		return false
	}

	return true
}
