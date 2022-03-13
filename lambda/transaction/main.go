package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"os"
	"transaction/aws"
	"transaction/parser"
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

var (
	nameAmountRegex       = os.Getenv("NameAmountRegex")
	moneyPoolsTableName   = os.Getenv("MoneyPoolsTableName")
	transactionsTableName = os.Getenv("TransactionsTableName")
)

func HandleRequest(_ context.Context, event EmailEvent) (string, error) {
	config := Config{
		ExpectedSubject: os.Getenv("EmailExpectedSubject"),
		MailGetter:      &aws.MailGetter{},
		MailParser:      parser.NewTransactionMailParser(nameAmountRegex),
		DataStore:       aws.NewDataStore(moneyPoolsTableName, transactionsTableName),
	}
	proc := NewMailEventProcessor(config)

	for _, record := range event.Records {
		proc.WriteTransactionToMoneyPool(record)
	}
	return "ok", nil
}

func main() {
	lambda.Start(HandleRequest)
}
