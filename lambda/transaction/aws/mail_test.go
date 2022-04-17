package aws

import (
	"bytes"
	"errors"
	"github.com/DusanKasan/parsemail"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
	"reflect"
	"testing"
)

type getMailTest struct {
	name          string
	getter        *MailGetter
	input         string
	expectedMail  *parsemail.Email
	expectedError error
}

func TestGetMail(t *testing.T) {

	testTable := []getMailTest{
		{
			"error",
			NewMailGetter(&ErrorDownloader{}),
			"messageId",
			nil,
			errors.New("Error while reading obj messageId from s3: error downloading file"),
		},
		{
			"valid",
			NewMailGetter(&ValidDownloader{}),
			"id",
			getMail("tests/test.mail"),
			nil,
		},
		{
			"tmp-write-error",
			NewMailGetter(&ValidDownloader{}),
			"",
			nil,
			errors.New("Error while reading obj  from s3: open /tmp/: is a directory"),
		},
	}
	for _, test := range testTable {
		mail, err := test.getter.GetMail(test.input)
		if !compareErrors(err, test.expectedError) || !reflect.DeepEqual(mail, test.expectedMail) {
			t.Fatalf("getMail(%s) = %v, %v but expected %v, %v", test.name, mail, err, test.expectedMail, test.expectedError)
		}
	}

}

type ErrorDownloader struct{}

func (v *ErrorDownloader) Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error) {
	return 0, errors.New("error downloading file")
}

type ValidDownloader struct{}

func (v *ValidDownloader) Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error) {
	mailFile, err := os.Open("tests/test.mail")
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, mailFile)
	if err != nil {
		panic(err)
	}
	_, err = w.WriteAt(buf.Bytes(), 0)
	if err != nil {
		panic(err)
	}
	return 0, nil
}

func compareErrors(err1, err2 error) bool {
	if err1 != nil && err2 != nil {
		return err1.Error() == err2.Error()
	}
	return err1 == err2
}

func getMail(filePath string) *parsemail.Email {
	mailFile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	mail, err := parsemail.Parse(mailFile)
	if err != nil {
		panic(err)
	}
	return &mail
}
