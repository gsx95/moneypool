package aws

import (
	"bytes"
	"fmt"
	"github.com/DusanKasan/parsemail"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
)

var (
	bucketName = os.Getenv("EmailBucketName")
)

type FileDownloader interface {
	Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error)
}

type MailGetter struct {
	FileDownloader
}

func NewMailGetter(downloader FileDownloader) *MailGetter {
	return &MailGetter{
		downloader,
	}
}

func (g *MailGetter) GetMail(messageId string) (mail *parsemail.Email, err error) {
	id := messageId
	contentReader, err := g.readFileFromS3(id)
	if err != nil {
		return nil, fmt.Errorf("Error while reading obj %s from s3: %v", id, err)
	}
	email, err := parsemail.Parse(contentReader)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing email of object %s from s3: %v", id, err)
	}
	return &email, nil
}

func (g *MailGetter) readFileFromS3(key string) (io.Reader, error) {
	file, err := os.Create("/tmp/" + key)
	if err != nil {
		return nil, err
	}

	_, err = g.FileDownloader.Download(file,
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
