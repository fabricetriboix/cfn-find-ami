package main

import (
	"context"
	"fmt"
	"errors"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/google/uuid"
)

func getString(event *cfn.Event, key string) (value string) {
	value = ""
	defer func() {
		recover()
	}()
	value = event.ResourceProperties[key].(string)
	return
}

func findAmi(ctx context.Context, event cfn.Event) (physicalId string, data map[string]interface{}, err error) {
	physicalId = uuid.New().String()  // Random UUID
	if event.RequestType == "Delete" {
		fmt.Println("Request type is 'Delete'; nothing to do")
		return
	}
	fmt.Printf("Request type is '%v'\n", event.RequestType)

	region := getString(&event, "Region")
	if region == "" {
		errorString := "'Region' must be set in the request's ResourceProperties"
		fmt.Println(errorString)
		err = errors.New(errorString)
		return
	}
	architecture := getString(&event, "Architecture")
	name := getString(&event, "Name")
	ownerId := getString(&event, "OwnerId")
	fmt.Println("Requested region:", region)
	fmt.Println("Requested architecture:", architecture)
	fmt.Println("Requested name:", name)
	fmt.Println("Requested owner id:", ownerId)

	cfg := &aws.Config{
		Region: &region,
	}
	sess := session.Must(session.NewSession(cfg))
	ec2Service := ec2.New(sess)

	input := new(ec2.DescribeImagesInput)
	if architecture != "" {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name: aws.String("architecture"),
			Values: []*string{&architecture},
		})
	}
	if name != "" {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name: aws.String("name"),
			Values: []*string{&name},
		})
	}
	if ownerId != "" {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name: aws.String("owner-id"),
			Values: []*string{&ownerId},
		})
	}

	output, err := ec2Service.DescribeImages(input)
	if err == nil {
		fmt.Println("API call to EC2.DescribeImages() succeeded")
		fmt.Printf("  %d images found", len(output.Images))
		fmt.Printf("  output: %+v\n", output)
		data = map[string]interface{} {
			"Id": output.Images[0].ImageId,
			"Name": output.Images[0].Name,
		}
	} else {
		fmt.Printf("API call to EC2.DescribeImages() failed: %v\n", err)
	}
	return
}

func main() {
	lambda.Start(cfn.LambdaWrap(findAmi))
}
