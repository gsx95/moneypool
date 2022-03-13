package main

import (
	"fmt"
	"github.com/DusanKasan/parsemail"
	"github.com/sirupsen/logrus"
	"os"
	"pay-event/aws"
	"pay-event/transaction"
	"strings"
	"time"
)

var (
	expectedSubject = os.Getenv("EmailExpectedSubject")
	log             = logrus.New()
)

func handleEventRecord(record EmailEventRecord) {
	log = log.WithFields(logrus.Fields{"messageId": record.Ses.Mail.MessageId}).Logger
	log.Infof("processing record")

	email, err := aws.ParseIncomingMail(record.Ses.Mail.MessageId)
	if err != nil {
		log.Errorf("error while parsing mail: %v", err)
		return
	}

	if strings.ToLower(email.Subject) != strings.ToLower(expectedSubject) {
		log.Infof("subject %s not matching expected %s", email.Subject, expectedSubject)
		return
	}

	transactionInfo, err := getTransactionInfoFromMail(*email)
	if err != nil {
		log.Errorf("error getting transaction info from mail: %v", err)
		return
	}

	moneyPools, err := findMoneyPoolsByPrefix(transactionInfo.Note)
	if err != nil {
		log.Errorf("error finding moneypool: %v", err)
		return
	}
	if len(moneyPools) > 1 {
		log.Errorf("ambiguous note, found multiple moneypools: %v", moneyPools)
		return
	}
	if len(moneyPools) < 1 {
		log.Infof("no moneypools found")
		return
	}
	moneyPool := moneyPools[0]
	log = log.WithFields(logrus.Fields{"pool": moneyPool}).Logger

	err = addToMoneyPool(moneyPool, transactionInfo)
	if err != nil {
		log.Errorf("error adding transaction to moneypool: %v", err)
		return
	}
}

func getTransactionInfoFromMail(email parsemail.Email) (transaction.Info, error) {
	info, err := transaction.GetTransactionInfo(email)
	if err != nil {
		return transaction.Info{}, fmt.Errorf("error while reading transaction infos form mail: %v", err)
	}
	log = log.WithFields(logrus.Fields{"sender": info.Name, "note": info.Note, "amount": info.AmountString()}).Logger
	log.Infof("found transaction info")
	return *info, err
}

func findMoneyPoolsByPrefix(note string) ([]string, error) {
	moneyPools, err := aws.FindMoneyPoolsByPrefix(note)
	if err != nil {
		return nil, fmt.Errorf("error while searching suitable moneypool: %v", err)
	}
	return moneyPools, nil
}

func addToMoneyPool(moneyPool string, transactionInfo transaction.Info) error {
	today := time.Now().Format("02.01.06")
	err := aws.AddTransaction(moneyPool, transactionInfo.Name, today, aws.Amount{
		Base:     transactionInfo.Base,
		Fraction: transactionInfo.Fraction,
	})
	if err != nil {
		return fmt.Errorf("error adding transaction to database: %v", err)
	}
	return nil
}
