This version of the Lambda function is written in Golang. It works
exactly the same as the Python version, except for building the
package, which is done like so:

```bash
$ go build find-ami.go
$ zip find-ami.zip find-ami
$ aws --profile YOUR_PROFILE lambda create-function \
    --function-name find-ami --memory 128 --timeout 30 \
    --role arn:aws:iam::123456789012:role/find-ami-lambda-role \
    --runtime go1.x --zip-file fileb://./find-ami.zip \
    --handler find-ami
```
