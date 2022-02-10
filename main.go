/*
	Return the number of invocations for the first 50 lambdas in an AWS region.
*/

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var (
	// Sess        *session.Session
	profileFlag string
	envFlag     bool
	regionFlag  string
	debugFlag   bool
	regions     []string
)

func getLambdaStats() {
	fmt.Println("Lambda invocation stats for last 28 days")
	fmt.Println("----------------------------------------")
	fmt.Println()
	if regionFlag == "all" {
		for i := 0; i < len(regions); i++ {
			getLambdaStatsByRegion(regions[i])
		}
	} else {
		getLambdaStatsByRegion(regionFlag)
	}
}

func getLambdaStatsByRegion(region string) {
	lambdas, err := listLambdas(region)
	if err != nil {
		log.Printf("Could not get lambdas in %v\n\n", region)
		if debugFlag {
			log.Printf("%v\n", err)
		}
		return
	}

	fmt.Printf("%v Lambda functions found in %v\n", len(lambdas.Functions), region)

	if len(lambdas.Functions) == 0 {
		fmt.Println()
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Function", "Type", "Invocations"})

	for i := 0; i < len(lambdas.Functions); i++ {
		/*
			metricsList, err := listMetrics(*lambdas.Functions[i].FunctionName)
			if err != nil {
				log.Printf("Could not get metrics: %v", err)
			}

			fmt.Printf("%v", metricsList)
		*/
		metricData, err := getMetricData(*lambdas.Functions[i].FunctionName, region)
		if err != nil {
			log.Printf("Could not get metrics: %v", err)
			return
		}

		var invocationCount float64

		if len(metricData.MetricDataResults) > 0 {
			if len(metricData.MetricDataResults[0].Values) > 0 {
				invocationCount = *metricData.MetricDataResults[0].Values[0]
			}
		}

		t.AppendRows([]table.Row{
			{*lambdas.Functions[i].FunctionName, *lambdas.Functions[i].Runtime, invocationCount},
		})
	}

	t.Render()
	fmt.Println()
}

func getMetricData(functionName, region string) (*cloudwatch.GetMetricDataOutput, error) {
	svc := cloudwatch.New(Session(region))

	dimensions := []*cloudwatch.Dimension{
		{
			Name:  aws.String("FunctionName"),
			Value: aws.String(functionName),
		},
	}

	metricStat := &cloudwatch.MetricStat{
		Metric: &cloudwatch.Metric{
			Dimensions: dimensions,
			Namespace:  aws.String("AWS/Lambda"),
			MetricName: aws.String("Invocations"),
		},
		Period: aws.Int64(86400 * 28), // 28 days
		Stat:   aws.String("Sum"),
		Unit:   aws.String("Count"),
	}

	ID := "invocations"

	queries := []*cloudwatch.MetricDataQuery{
		{
			Id:         &ID,
			MetricStat: metricStat,
		},
	}

	input := &cloudwatch.GetMetricDataInput{}
	input.SetMetricDataQueries(queries)
	input.SetEndTime(time.Now())
	input.SetStartTime(time.Now().AddDate(-0, -0, -28))

	result, err := svc.GetMetricData(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudwatch.ErrCodeInvalidNextToken:
				return nil, errors.New(fmt.Sprintln(cloudwatch.ErrCodeInvalidNextToken, aerr.Error()))
			default:
				return nil, errors.New(fmt.Sprintln(aerr.Error()))
			}
		} else {
			return nil, err
		}
	}

	return result, nil
}

/*
func listMetrics(functionName string) (*cloudwatch.ListMetricsOutput, error) {
	svc := cloudwatch.New(Session())

	dimensions := []*cloudwatch.DimensionFilter{
		{
			Name:  aws.String("FunctionName"),
			Value: aws.String(functionName),
		},
	}

	input := &cloudwatch.ListMetricsInput{
		Dimensions: dimensions,
		Namespace:  aws.String("AWS/Lambda"),
		MetricName: aws.String("Invocations"),
	}

	result, err := svc.ListMetrics(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudwatch.ErrCodeInternalServiceFault:
				return nil, errors.New(fmt.Sprintln(cloudwatch.ErrCodeInternalServiceFault, aerr.Error()))
			case cloudwatch.ErrCodeInvalidParameterValueException:
				return nil, errors.New(fmt.Sprintln(cloudwatch.ErrCodeInvalidParameterValueException, aerr.Error()))
			default:
				return nil, errors.New(fmt.Sprintln(aerr.Error()))
			}
		} else {
			return nil, err
		}
	}

	return result, nil
}
*/

func listLambdas(region string) (*lambda.ListFunctionsOutput, error) {
	svc := lambda.New(Session(region))
	input := &lambda.ListFunctionsInput{}

	result, err := svc.ListFunctions(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				return nil, errors.New(fmt.Sprintln(lambda.ErrCodeServiceException, aerr.Error()))
			case lambda.ErrCodeTooManyRequestsException:
				return nil, errors.New(fmt.Sprintln(lambda.ErrCodeTooManyRequestsException, aerr.Error()))
			case lambda.ErrCodeInvalidParameterValueException:
				return nil, errors.New(fmt.Sprintln(lambda.ErrCodeInvalidParameterValueException, aerr.Error()))
			default:
				return nil, errors.New(fmt.Sprintln(aerr.Error()))
			}
		} else {
			return nil, err
		}
	}

	return result, nil
}

func Session(region string) *session.Session {
	verboseCredentialErrors := true

	sessionOpts := session.Options{
		// Specify profile to load for the session's config
		// Profile: profile,

		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region:                        aws.String(region),
			CredentialsChainVerboseErrors: &verboseCredentialErrors,
		},

		// Force enable Shared Config support
		// SharedConfigState: session.SharedConfigEnable,
	}

	if !envFlag {
		sessionOpts.Profile = profileFlag
	}

	sess, err := session.NewSessionWithOptions(sessionOpts)
	if err != nil {
		fmt.Printf("%v", err)
		return nil
	}

	return sess
}

func main() {
	regions = []string{
		"af-south-1",
		"ap-east-1",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-south-1",
		"ap-southeast-3",
		"ca-central-1",
		"eu-central-1",
		"eu-north-1",
		"eu-south-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"me-south-1",
		"sa-east-1",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}

	var rootCmd = &cobra.Command{
		Use:   "lambda-invocations",
		Short: "List the number of invocations per AWS lambda in a region",
		Run: func(cmd *cobra.Command, args []string) {
			getLambdaStats()
		},
	}

	rootCmd.Flags().StringVarP(&profileFlag, "profile", "p", "default", "AWS credentials profile name")
	rootCmd.Flags().BoolVarP(&envFlag, "env", "e", false, "Use environment vars for credentials")
	rootCmd.Flags().StringVarP(&regionFlag, "region", "r", "eu-west-2", "AWS region (e.g. us-east-1). Use \"all\" for all regions")
	rootCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "Debug")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
