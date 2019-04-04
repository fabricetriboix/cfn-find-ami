cfn-find-ami
============

AWS Lambda function to find AMIs, to be used as a custom resource in
CloudFormation templates.

Install the lambda function
---------------------------

Create an IAM policy using the `find-ami-lambda-policy.json` file.
Create an IAM role for the lambda function using the above policy.

Then run the following commands in a bash shell:

```bash
$ go build find-ami.go
$ zip find-ami.zip find-ami
$ aws --region YOUR-REGION lambda create-function \
    --function-name find-ami --memory 128 \
    --role arn:aws:iam::281582310019:role/find-ami-lambda-role \
    --runtime go1.x --zip-file fileb://./find-ami.zip \
    --handler find-ami
```

"Call" the lambda function from a CloudFormation template
---------------------------------------------------------

Checkout the `test-cfn-find-ami.yml` for an example.

The following `Properties` can be set in the custom resource:
 - `ServiceToken`: Set to the ARN of the above lambda function
 - `Region`: The region you want to search AMIs in; this property is
   mandatory
 - `Debug`: Set to `true` to increase verbosity (the logs from the
   lambda functions are available in the CloudWatch logs)
 - `Architecture`: The architecture to filter; mostly likely you
   should set this to `x86_64`
 - `Name`: Filter on the AMI names; you can use `*` as wildcards;
   for example, `*bionic*` will search for images with "bionic"
   anywhere in their names
 - `OwnerId`: Filter on the owner id
 - `RootDeviceType`: Filter on the root device type; can be "ebs" or
   "instance-store"; if not set, this defaults to "ebs"
 - `VirtualizationType`: Filter on the virtualization type; can be
   "hvm" or "paravirtual"; if not set, this defaults to "hvm"

The lambda function will filter AMIs that match the parameters you
provided, and will return the most recent image. The following output
parameters are available:
 - `Id`: The AMI id of the found AMI
 - `Name`: The name of the found AMI
 - `Description`: The description of the found AMI
