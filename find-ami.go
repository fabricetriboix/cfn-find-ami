package main


import (
	"context"
	"fmt"
	"errors"
	"sort"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/google/uuid"
)


// Extract a string from the given `event`. Returns the empty string
// in case of any error.
func getString(event *cfn.Event, key string) (value string) {
	value = ""
	defer func() {
		recover()
	}()
	value = event.ResourceProperties[key].(string)
	return
}


// Lambda event handler
func findAmi(ctx context.Context, event cfn.Event) (physicalId string, data map[string]interface{}, err error) {
	if event.RequestType == "Delete" {
		fmt.Println("Request type is 'Delete'; nothing to do")
		return
	}
	fmt.Printf("Request type is '%v'\n", event.RequestType)

	// Generate a random physial ID to keep CloudFormation happy
	physicalId = uuid.New().String()  // Random UUID

	// Extract filter values from the `ResourceProperties`
	region := getString(&event, "Region")
	if region == "" {
		err = errors.New("'Region' must be set in the request's ResourceProperties")
		fmt.Println(err)
		return
	}
	debug := getString(&event, "Debug")
	architecture := getString(&event, "Architecture")
	name := getString(&event, "Name")
	ownerId := getString(&event, "OwnerId")
	rootDeviceType := getString(&event, "RootDeviceType")
	if rootDeviceType == "" {
		rootDeviceType = "ebs"
	}
	virtualizationType := getString(&event, "VirtualizationType")
	if virtualizationType == "" {
		virtualizationType = "hvm"
	}
	if debug == "true" {
		fmt.Println("Requested debug: true")
		fmt.Println("Requested region:", region)
		fmt.Println("Requested architecture:", architecture)
		fmt.Println("Requested name:", name)
		fmt.Println("Requested owner id:", ownerId)
		fmt.Println("Requested root device type:", rootDeviceType)
	}

	// Query the EC2 service to list the matching AMIs
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
	input.Filters = append(input.Filters, &ec2.Filter{
		Name: aws.String("root-device-type"),
		Values: []*string{&rootDeviceType},
	})
	input.Filters = append(input.Filters, &ec2.Filter{
		Name: aws.String("virtualization-type"),
		Values: []*string{&virtualizationType},
	})
	output, err := ec2Service.DescribeImages(input)
	if err != nil {
		fmt.Printf("API call to EC2.DescribeImages() failed: %v\n", err)
		return
	}
	fmt.Println("API call to EC2.DescribeImages() succeeded")
	fmt.Printf("%d images found", len(output.Images))
	if len(output.Images) == 0 {
		errorString := "No image found for the given filters"
		err = errors.New(errorString)
		return
	}

	// Sort the matching AMIs by creation date and return the latest one
	sort.Slice(output.Images, func(i, j int) bool {
		return *(output.Images[i].CreationDate) > *(output.Images[j].CreationDate)
	})
	data = map[string]interface{} {
		"Id": *(output.Images[0].ImageId),
		"Name": *(output.Images[0].Name),
		"Description": *(output.Images[0].Description),
	}
	return
}


func main() {
	lambda.Start(cfn.LambdaWrap(findAmi))
}
