package aws

import (
	"bytes"
	"fmt"
	"github.com/DusanKasan/parsemail"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"log"
	"os"
	"strings"
)

var (
	expectedSubject = os.Getenv("EmailExpectedSubject")

	Session    = session.Must(session.NewSession())
	downloader = s3manager.NewDownloader(Session)
	bucketName = os.Getenv("EmailBucketName")
)

func ParseMoneyPoolMail(messageId string) (mail *parsemail.Email, err error) {
	id := messageId
	contentReader, err := readFileFromS3(id)
	if err != nil {
		return nil, fmt.Errorf("Error while reading obj %s from s3: %v\n", id, err)
	}
	email, err := parsemail.Parse(contentReader)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing email of object %s from s3: %v\n", id, err)
	}
	log.Printf("got aws %s with subject %s\n", id, email.Subject)
	if strings.ToLower(email.Subject) != strings.ToLower(expectedSubject) {
		return nil, fmt.Errorf("subject %s not matching expcted %s", email.Subject, expectedSubject)
	}
	return &email, nil
}

func readFileFromS3(key string) (io.Reader, error) {
	file, err := os.Create("/tmp/" + key)
	if err != nil {
		return nil, err
	}

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}
