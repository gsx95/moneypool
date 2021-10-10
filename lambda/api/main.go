package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"os"
	"strconv"
)

var (
	moneyPoolsTableName   = os.Getenv("MoneyPoolsTableName")
	transactionsTableName = os.Getenv("TransactionsTableName")
	corsDomain            = os.Getenv("CorsDomain")
	awsSession            = session.Must(session.NewSession())
	dynamoClient          = dynamodb.New(awsSession, aws.NewConfig())
)

type Transaction struct {
	Name     string `json:"name"`
	Base     int    `json:"base"`
	Fraction int    `json:"fraction"`
}

type Response struct {
	Transactions []Transaction `json:"transactions"`
	Name         string        `json:"name"`
	Title        string        `json:"title"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	mpName := request.PathParameters["moneyPool"]

	log.Println("getting details for MP " + mpName)

	mpItem, err := dynamoClient.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(mpName),
			},
		},
		TableName: aws.String(moneyPoolsTableName),
	})

	if err != nil {
		log.Println(err)
		return notFoundResponse(), nil
	}

	resp := &Response{}

	resp.Name = *mpItem.Item["name"].S
	resp.Title = *mpItem.Item["title"].S

	for _, tValues := range mpItem.Item["transactions"].L {
		name, base, fraction, err := getTransaction(*tValues.S)
		if err != nil {
			log.Printf("Error getting transaction %s: %v", *tValues.S, err)
			continue
		}
		resp.Transactions = append(resp.Transactions, Transaction{
			Name:     name,
			Base:     base,
			Fraction: fraction,
		})
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error while marshalling response %v: %v", resp, err)
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

func getTransaction(tid string) (name string, base, fraction int, err error) {
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
	baseString := *trItem.Item["base"].N
	fractionString := *trItem.Item["fraction"].N
	log.Printf("got transaction: %s %s %s", name, baseString, fractionString)
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

func notFoundResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "Not found",
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
