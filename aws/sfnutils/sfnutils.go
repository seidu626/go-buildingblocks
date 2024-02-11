package sfnutils

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awsutils "github.com/seidu626/go-buildingblocks/aws"
	"sync"
)

var sfnClient *sfn.Client = nil
var once sync.Once

func init() {
	once.Do(func() {
		cfg, err := awsutils.New()
		if err != nil {
			panic(err)
		}
		sfnClient = sfn.New(sfn.Options{Credentials: cfg.Credentials, Region: cfg.Region, RetryMaxAttempts: 5, RetryMode: aws.RetryModeAdaptive})
	})
}

func StartSFN(sfnArn, payload string) (*sfn.StartExecutionOutput, error) {
	return sfnClient.StartExecution(context.Background(), &sfn.StartExecutionInput{
		StateMachineArn: aws.String(sfnArn),
		Input:           aws.String(payload),
	})
}
