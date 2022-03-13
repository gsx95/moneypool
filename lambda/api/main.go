package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

var (
	moneyPoolsTableName   = os.Getenv("MoneyPoolsTableName")
	transactionsTableName = os.Getenv("TransactionsTableName")
	corsDomain            = os.Getenv("CorsDomain")
	awsSession            = session.Must(session.NewSession())
	dynamoClient          = dynamodb.New(awsSession, aws.NewConfig())

	logger = log.New()
)

type Transaction struct {
	Name     string `json:"name"`
	Base     int    `json:"base"`
	Date     string `json:"date,omitempty"`
	Fraction int    `json:"fraction"`
}

type Response struct {
	Transactions []Transaction `json:"transactions"`
	Name         string        `json:"name"`
	Title        string        `json:"title"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	mpName, mpParamExists := request.PathParameters["moneyPool"]
	if !mpParamExists {
		log.Errorf("no moneypool name given")
		return badRequestResponse(), nil
	}

	logger = logger.WithFields(log.Fields{"requestedMP": mpName}).Logger
	logger.Infof("search moneyppol")
	mpItem, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(mpName),
			},
		},
		TableName: aws.String(moneyPoolsTableName),
	})
	if err != nil {
		logger.Errorf("error getting moneypool from db: %v", err)
		return notFoundResponse(), nil
	}

	items := mpItem.Item

	if _, exists := items["name"]; !exists {
		logger.Errorf("moneypool item has no name field")
		return internalErrorResponse(), nil
	}

	if _, exists := items["title"]; !exists {
		logger.Errorf("moneypool item has no title field")
		return internalErrorResponse(), nil
	}

	resp := &Response{
		Name:  *mpItem.Item["name"].S,
		Title: *mpItem.Item["title"].S,
	}

	for _, tValues := range mpItem.Item["transactions"].L {
		name, date, base, fraction, err := getTransaction(*tValues.S)
		if err != nil {
			log.Errorf("Error getting transaction %s: %v", *tValues.S, err)
			continue
		}
		resp.Transactions = append(resp.Transactions, Transaction{
			Name:     name,
			Date:     date,
			Base:     base,
			Fraction: fraction,
		})
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("Error while marshalling response %v: %v", resp, err)
		return internalErrorResponse(), nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(jsonResp),
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,GET",
		},
	}, nil
}

func getTransaction(tid string) (name, date string, base, fraction int, err error) {
	trItem, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(tid),
			},
		},
		TableName: aws.String(transactionsTableName),
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
	log.Infof("got transaction: %s %s %s", name, baseString, fractionString)
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

func badRequestResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "bad request",
		StatusCode: 400,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,GET",
		},
	}
}

func notFoundResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "not found",
		StatusCode: 404,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,GET",
		},
	}
}

func internalErrorResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "internal error",
		StatusCode: 500,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,GET",
		},
	}
}

func main() {
	lambda.Start(handler)
}
