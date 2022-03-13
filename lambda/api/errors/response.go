package errors

import (
	er "errors"
	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"
)

var (
	invalidParamsError *InvalidParametersError
	notFoundError      *NotFoundError
)

func ToResponse(err error) events.APIGatewayProxyResponse {
	if er.As(err, &invalidParamsError) {
		log.Errorf("Invalid Parameter Error: %v", err)
		return badRequestResponse()
	}

	if er.As(err, &notFoundError) {
		log.Errorf("Invalid Parameter Error: %v", err)
		return notFoundResponse()
	}
	log.Errorf("internal Error: %v", err)
	return internalErrorResponse()
}

func badRequestResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "bad request",
		StatusCode: 400,
	}
}

func notFoundResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "not found",
		StatusCode: 404,
	}
}

func internalErrorResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "internal error",
		StatusCode: 500,
	}
}
