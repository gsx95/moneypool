package main

import (
	"api/errors"
	"api/moneypool"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"os"
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
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	poolsHandler := moneypool.NewHandler(moneyPoolsTableName, transactionsTableName, corsDomain, *dynamoClient)
	moneyPool, err := poolsHandler.GetMoneyPool(request)

	if err != nil {
		return addHeaderToResponse(errors.ToResponse(err)), nil
	}

	jsonResp, err := json.Marshal(moneyPool)
	if err != nil {
		err = fmt.Errorf("error while marshalling response %v: %v", moneyPool, err)
		return addHeaderToResponse(errors.ToResponse(err)), nil
	}

	return addHeaderToResponse(events.APIGatewayProxyResponse{
		Body:       string(jsonResp),
		StatusCode: 200,
	}), nil
}

func addHeaderToResponse(response events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	response.Headers = map[string]string{
		"Access-Control-Allow-Headers": "*",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "OPTIONS,GET",
	}
	return response
}

func main() {
	lambda.Start(handler)
}
