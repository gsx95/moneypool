package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"github.com/DusanKasan/parsemail"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type EmailEvent struct {
	Records []EmailEventRecord `json:"Records"`
}

type EmailEventRecord struct {
	Ses struct {
		Mail struct {
			Timestamp string `json:"timestamp"`
			MessageId string `json:"messageId"`
		} `json:"mail"`
	} `json:"ses"`
}

var (
	AwsSession = session.Must(session.NewSession())
	downloader = s3manager.NewDownloader(AwsSession)

	bucketName            = os.Getenv("EmailBucketName")
	expectedSubject       = os.Getenv("EmailExpectedSubject")
	notificationTextRegex = os.Getenv("EmailNotificationTextRegex")
	nameAmountSeparator   = os.Getenv("EmailNameAmountSeparator")
	amountTextSeparator   = os.Getenv("EmailAmountTextSeparator")
	decimalSeparator      = os.Getenv("CurrencyDecimalSeparator")
	thousandsSeparator    = os.Getenv("CurrencyThousandsSeparator")
)

func HandleRequest(ctx context.Context, event EmailEvent) (string, error) {

	log.Println(event)

	for _, record := range event.Records {
		id := record.Ses.Mail.MessageId
		contentReader, err := readFileFromS3(id)

		if err != nil {
			log.Printf("Error while reading obj %s from s3: %v\n", id, err)
			continue
		}

		email, err := parsemail.Parse(contentReader)
		if err != nil {
			log.Printf("Error while parsing email of object %s from s3: %v\n", id, err)
			continue
		}
		log.Printf("got mail %s with subject %s\n", id, email.Subject)
		if strings.ToLower(email.Subject) != strings.ToLower(expectedSubject) {
			continue
		}

		decodedHtml, err := b64.StdEncoding.DecodeString(email.HTMLBody)
		if err != nil {
			log.Printf("Error while decoding base64 html %v\n", err)
			continue
		}
		notificationText := regexp.MustCompile(notificationTextRegex).Find(decodedHtml)
		tokens := strings.Split(string(notificationText), nameAmountSeparator)
		name := tokens[0]
		amount := tokens[1][:strings.Index(tokens[1], amountTextSeparator)]
		amountSplit := strings.Split(amount, decimalSeparator)

		fraction, err := strconv.Atoi(amountSplit[1])
		if err != nil {
			log.Printf("Error while parsing fraction of %s: %v\n", amount, err)
			continue
		}
		base, err := strconv.Atoi(strings.Replace(amountSplit[0], thousandsSeparator, "", -1))
		if err != nil {
			log.Printf("Error while parsing base of %s: %v\n", amount, err)
			continue
		}

		htmlBetweenQuotes := substringBetween(string(decodedHtml), "quote-left.png", "quote-right.png")
		tagWithNoteText := substringBetween(htmlBetweenQuotes, `align="center"`, "</td>")
		note := tagWithNoteText[strings.Index(tagWithNoteText, ">")+1:]
		note = note[strings.Index(note, ">")+1:]
		note = note[strings.Index(note, ">")+1:]
		note = note[:strings.Index(note, "<")]

		allMoneyPools := GetAllMoneyPools()
		var moneyPool *string
		for _, mpName := range allMoneyPools {
			if strings.HasPrefix(strings.ToLower(note), strings.ToLower(*mpName)) {
				moneyPool = mpName
				break
			}
		}

		if moneyPool == nil {
			log.Printf("no moneypool matches note %s\n", note)
			continue
		}

		err = AddTransaction(*moneyPool, name, Amount{
			Base:     base,
			Fraction: fraction,
		})
		if err != nil {
			log.Printf("Error while adding transaction for %s and %s: %v\n", *moneyPool, name, err)
			continue
		}

	}
	return "ok", nil
}

func readFileFromS3(key string) (io.Reader, error) {
	file, err := os.Create("/tmp/" + key)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error creating temp file: %v", err))
	}

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error download obj into file: %v", err))
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	return bytes.NewReader(buf.Bytes()), nil
}

func substringBetween(str, start, end string) string {
	startIndex := strings.Index(str, start) + len(start)
	quoteRightIndex := strings.Index(str[startIndex:], end) + startIndex - 1
	return str[startIndex:quoteRightIndex]
}

func main() {
	lambda.Start(HandleRequest)
}
