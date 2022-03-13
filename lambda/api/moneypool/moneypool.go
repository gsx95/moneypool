package moneypool

import (
	"api/errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type Transaction struct {
	Name     string `json:"name"`
	Base     int    `json:"base"`
	Date     string `json:"date,omitempty"`
	Fraction int    `json:"fraction"`
}

type MoneyPool struct {
	Transactions []Transaction `json:"transactions"`
	Name         string        `json:"name"`
	Title        string        `json:"title"`
}

type MoneyPoolsHandler struct {
	moneyPoolsTableName   string
	transactionsTableName string
	corsDomain            string
	dynamoClient          dynamodb.DynamoDB
	logger                *log.Entry
}

func NewHandler(moneyPoolsTableName string, transactionsTableName string, corsDomain string, dynamoClient dynamodb.DynamoDB) *MoneyPoolsHandler {
	return &MoneyPoolsHandler{moneyPoolsTableName: moneyPoolsTableName, transactionsTableName: transactionsTableName, corsDomain: corsDomain, dynamoClient: dynamoClient}
}

func (h *MoneyPoolsHandler) GetMoneyPool(request events.APIGatewayProxyRequest) (MoneyPool, error) {
	mpName, mpParamExists := request.PathParameters["moneyPool"]
	if !mpParamExists {
		return MoneyPool{}, errors.NewInvalidParametersError(fmt.Errorf("no moneyppol name given"))
	}

	h.logger = log.WithFields(log.Fields{"requestedMP": mpName})
	h.logger.Infof("search moneypool")
	mpItem, err := h.dynamoClient.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(mpName),
			},
		},
		TableName: aws.String(h.moneyPoolsTableName),
	})
	if err != nil {
		return MoneyPool{}, fmt.Errorf("error getting moneypool from db: %v", err)
	}
	h.logger.Infof("found moneypool item")
	item := mpItem.Item
	if item == nil {
		return MoneyPool{}, errors.NewNotFoundError(fmt.Errorf("no moneypool found for given name %s", mpName))
	}

	if _, exists := item["name"]; !exists {
		return MoneyPool{}, fmt.Errorf("moneypool item has no name field")
	}

	if _, exists := item["title"]; !exists {
		return MoneyPool{}, fmt.Errorf("moneypool item has no title field")
	}

	resp := MoneyPool{
		Name:  *mpItem.Item["name"].S,
		Title: *mpItem.Item["title"].S,
	}

	for _, tValues := range mpItem.Item["transactions"].L {
		name, date, base, fraction, err := h.getTransaction(*tValues.S)
		if err != nil {
			h.logger.Errorf("Error getting parser %s: %v", *tValues.S, err)
			continue
		}
		resp.Transactions = append(resp.Transactions, Transaction{
			Name:     name,
			Date:     date,
			Base:     base,
			Fraction: fraction,
		})
	}
	h.logger.Infof("moneypool item: %+v", resp)
	return resp, nil
}

func (h *MoneyPoolsHandler) getTransaction(tid string) (name, date string, base, fraction int, err error) {
	trItem, err := h.dynamoClient.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(tid),
			},
		},
		TableName: aws.String(h.transactionsTableName),
	})
	if err != nil {
		return
	}
	name = *trItem.Item["name"].S
	if trItem.Item["date"] != nil {
		date = *trItem.Item["date"].S
	}
	baseString := *trItem.Item["base"].N
	fractionString := *trItem.Item["fraction"].N
	h.logger.Infof("got parser: %s %s %s", name, baseString, fractionString)
	base, err = strconv.Atoi(baseString)
	if err != nil {
		return
	}
	fraction, err = strconv.Atoi(fractionString)
	if err != nil {
		return
	}
	return
}
