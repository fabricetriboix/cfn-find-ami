#!/usr/bin/env python3

import boto3
import uuid
import json
import requests


def handler(event, context):
    """
    Handle an operation for a "FindAmi" CloudFormation custom resource.

    The received event has the following pattern:

    {
       "RequestType" : "Create",
       "ResponseURL" : "http://pre-signed-S3-url-for-response",
       "StackId" : "arn:aws:cloudformation:us-west-2:123456789012:stack/stack-name/guid",
       "RequestId" : "unique id for this create request",
       "ResourceType" : "Custom::TestResource",
       "LogicalResourceId" : "MyTestResource",
       "ResourceProperties" : {
          "Name" : "Value",
          "List" : [ "1", "2", "3" ]
       }
    }
    """
    if event['RequestType'] == "Delete":
        print(f"Request type is 'Delete'; nothing to do")
        send_response(event, True)
        return
    print(f"Request type is '{event['RequestType']}'")

    prop = event['ResourceProperties']
    region = prop.get('Region')
    if not region:
        msg = f"'Region' must be set in the request's ResourceProperties"
        print(msg)
        send_response(event, False, msg)
        return

    debug = prop.get('Debug')
    if debug == "true":
        debug = True
    else:
        debug = False

    architecture = prop.get('Architecture', "")
    name = prop.get('Name', "")
    owner_id = prop.get('OwnerId', "")
    root_device_type = prop.get('RootDeviceType', "ebs")
    virtualization_type = prop.get('VirtualizationType', "hvm")

    if debug:
        print(f"Requested debug: true")
        print(f"Requested region: {region}")
        print(f"Requested architecture: {architecture}")
        print(f"Requested name: {name}")
        print(f"Requested owner id: {owner_id}")
        print(f"Requested root device type: {root_device_type}")
        print(f"Requested virtualization type: {virtualization_type}")

    ec2 = boto3.client("ec2", region_name=region)

    filters = []
    if architecture:
        filters.append({'Name': "architecture", 'Values': [architecture]})
    if name:
        filters.append({'Name': "name", 'Values': [name]})
    if owner_id:
        filters.append({'Name': "owner-id", 'Values': [owner_id]})
    filters.append({'Name': "root-device-type", 'Values': [root_device_type]})
    filters.append({'Name': "virtualization-type", 'Values': [virtualization_type]})

    response = ec2.describe_images(Filters=filters)
    images = response['Images']
    if not images:
        msg = f"No image found for filters: {filters}"
        print(f"ERROR: {msg}")
        send_response(event, False, msg)
        return
    print(f"API call to ec2.describe_images() succeeded; {len(images)} images found")

    # Sort the matching AMIs by creation date and return the latest one
    sorted_images = sorted(images, key=lambda i: i['CreationDate'], reverse=True)
    image = sorted_images[0]
    data = {
        'Id': image['ImageId'],
        'Name': image['Name'],
        'Description': image['Description']
    }
    print(f"Latest image: {data}")

    send_response(event, True, "Success", data)


def send_response(event, success: bool, msg="", data={}):
    response = {
        'Status': "SUCCESS" if success else "FAILED",
        'Reason': msg,
        'PhysicalResourceId': str(uuid.uuid4()),
        'StackId': event['StackId'],
        'RequestId': event['RequestId'],
        'LogicalResourceId': event['LogicalResourceId'],
        'Data': data
    }
    data = json.dumps(response)
    requests.put(event['ResponseURL'], data=data)
