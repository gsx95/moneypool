package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"process-payment-notification/aws"
	"process-payment-notification/transaction"
	"time"
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

func HandleRequest(_ context.Context, event EmailEvent) (string, error) {

	log.Println(event)
	for _, record := range event.Records {
		err := handleMailRecord(record)
		if err != nil {
			log.Println(err)
		}
	}
	return "ok", nil
}

func handleMailRecord(record EmailEventRecord) error {
	email, err := aws.ParseMoneyPoolMail(record.Ses.Mail.MessageId)
	if err != nil {
		return err
	}
	transactionInfo, err := transaction.GetTransactionInfo(*email)

	moneyPool, err := aws.FindMoneyPoolByName(transactionInfo.Note)
	if err != nil {
		return err
	}
	today := time.Now().Format("02.01.06")
	err = aws.AddTransaction(moneyPool, transactionInfo.Name, today, aws.Amount{
		Base:     transactionInfo.Base,
		Fraction: transactionInfo.Fraction,
	})
	if err != nil {
		return fmt.Errorf("Error while adding transaction for %s and %s: %v\n", moneyPool, transactionInfo.Name, err)
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
