package main

import (
	"fmt"
	"github.com/DusanKasan/parsemail"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"transaction/data"
)

type MailParser interface {
	GetTransactionInfo(email parsemail.Email) (*data.Transaction, error)
}

type MailGetter interface {
	GetMail(messageId string) (mail *parsemail.Email, err error)
}

type DataStore interface {
	FindMoneyPoolsByPrefix(name string) ([]string, error)
	AddTransaction(moneyPool, name, date string, amount data.Amount) error
}

type Config struct {
	ExpectedSubject string
	MailParser      MailParser
	MailGetter      MailGetter
	DataStore       DataStore
}

type MailEventProcessor struct {
	Config
	logger *logrus.Logger
}

func NewMailEventProcessor(config Config) MailEventProcessor {
	return MailEventProcessor{
		Config: config,
	}
}

func (h *MailEventProcessor) WriteTransactionToMoneyPool(record EmailEventRecord) {
	h.logger = logrus.New()
	h.logger = h.logger.WithFields(logrus.Fields{"messageId": record.Ses.Mail.MessageId}).Logger
	h.logger.Infof("processing record")

	email, err := h.MailGetter.GetMail(record.Ses.Mail.MessageId)
	if err != nil {
		h.logger.Errorf("error while parsing mail: %v", err)
		return
	}

	if strings.ToLower(email.Subject) != strings.ToLower(h.ExpectedSubject) {
		h.logger.Infof("subject %s not matching expected %s", email.Subject, h.ExpectedSubject)
		return
	}

	transactionInfo, err := h.getTransactionInfoFromMail(*email)
	if err != nil {
		h.logger.Errorf("error getting parser info from mail: %v", err)
		return
	}

	moneyPools, err := h.findMoneyPoolsByPrefix(transactionInfo.Note)
	if err != nil {
		h.logger.Errorf("error finding moneypool: %v", err)
		return
	}
	if len(moneyPools) > 1 {
		h.logger.Errorf("ambiguous note, found multiple moneypools: %v", moneyPools)
		return
	}
	if len(moneyPools) < 1 {
		h.logger.Infof("no moneypools found")
		return
	}
	moneyPool := moneyPools[0]
	h.logger = h.logger.WithFields(logrus.Fields{"pool": moneyPool}).Logger

	err = h.addToMoneyPool(moneyPool, transactionInfo)
	if err != nil {
		h.logger.Errorf("error adding parser to moneypool: %v", err)
		return
	}
}

func (h *MailEventProcessor) getTransactionInfoFromMail(email parsemail.Email) (data.Transaction, error) {
	info, err := h.MailParser.GetTransactionInfo(email)
	if err != nil {
		return data.Transaction{}, fmt.Errorf("error while reading parser infos form mail: %v", err)
	}
	h.logger = h.logger.WithFields(logrus.Fields{"sender": info.Name, "note": info.Note, "amount": fmt.Sprintf("%d.%d", info.Base, info.Fraction)}).Logger
	h.logger.Infof("found parser info")
	return *info, err
}

func (h *MailEventProcessor) findMoneyPoolsByPrefix(note string) ([]string, error) {
	moneyPools, err := h.DataStore.FindMoneyPoolsByPrefix(note)
	if err != nil {
		return nil, fmt.Errorf("error while searching suitable moneypool: %v", err)
	}
	return moneyPools, nil
}

func (h *MailEventProcessor) addToMoneyPool(moneyPool string, transactionInfo data.Transaction) error {
	today := time.Now().Format("02.01.06")
	err := h.DataStore.AddTransaction(moneyPool, transactionInfo.Name, today, data.Amount{
		Base:     transactionInfo.Base,
		Fraction: transactionInfo.Fraction,
	})
	if err != nil {
		return fmt.Errorf("error adding parser to database: %v", err)
	}
	return nil
}
