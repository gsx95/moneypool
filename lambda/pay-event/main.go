package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"os"
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

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

func HandleRequest(_ context.Context, event EmailEvent) (string, error) {
	for _, record := range event.Records {
		handleEventRecord(record)
	}
	return "ok", nil
}

func main() {
	lambda.Start(HandleRequest)
}
