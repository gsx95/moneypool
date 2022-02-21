package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	moneyPoolsTableName   = os.Getenv("MoneyPoolsTableName")
	transactionsTableName = os.Getenv("TransactionsTableName")
	dynamoClient          = dynamodb.New(Session, aws.NewConfig())
)

type Amount struct {
	Base     int // e.g. eur, usd
	Fraction int // e.g. cents
}

func FindMoneyPoolByName(name string) (string, error) {
	allMoneyPools := getAllMoneyPools()
	for _, mpName := range allMoneyPools {
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(*mpName)) {
			return *mpName, nil
		}
	}
	return "", fmt.Errorf("no moneypool found with name %s", name)
}

func getAllMoneyPools() (names []*string) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(moneyPoolsTableName),
		AttributesToGet: []*string{
			aws.String("name"),
		},
	}
	results, err := dynamoClient.Scan(input)
	if err != nil {
		log.Printf("Could not get all money pools %v \n", err)
		return
	}

	for _, result := range results.Items {
		name := result["name"]
		names = append(names, name.S)
	}
	return
}

func AddTransaction(moneyPool, name, date string, amount Amount) error {

	uid := uuid.New().String()

	tIds := []*dynamodb.AttributeValue{
		{
			S: aws.String(uid),
		},
	}

	log.Printf("add transaction: %s %s", moneyPool, uid)

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(moneyPool),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":tid": {
				L: tIds,
			},
			":empty_list": {
				L: []*dynamodb.AttributeValue{},
			},
		},
		UpdateExpression: aws.String("SET transactions = list_append(if_not_exists(transactions, :empty_list), :tid)"),
		TableName:        aws.String(moneyPoolsTableName),
	}

	_, err := dynamoClient.UpdateItem(input)
	if err != nil {
		return errors.New(fmt.Sprintf("Error updating moneypool item: %v\n", err))
	}

	log.Printf("add transaction: %s %s - %d - %d", moneyPool, name, amount.Base, amount.Fraction)

	transactionInput := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(uid),
			},
			"base": {
				N: aws.String(strconv.Itoa(amount.Base)),
			},
			"fraction": {
				N: aws.String(strconv.Itoa(amount.Fraction)),
			},
			"name": {
				S: aws.String(name),
			},
			"date": {
				S: aws.String(date),
			},
		},
		TableName: aws.String(transactionsTableName),
	}

	_, err = dynamoClient.PutItem(transactionInput)
	if err != nil {
		return errors.New(fmt.Sprintf("Error putting transaction item: %v\n", err))
	}
	return nil
}