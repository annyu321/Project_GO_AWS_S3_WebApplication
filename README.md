# GO-AWS-S3-WebApplication
## Introduction
This GO AWS S3 Web Application provides a variety of APIs to perform S3 bucket operations using AWS S3 SDK v1 and Go programming language. It uses the LoggingHandler middleware from the gorilla/handlers package to write logs to a file that records requests as the Apache Common Log Format.

## Web APIs
### Create
#### Create a new S3 bucket named as "album-Year-Month-Date"
#### http://localhost:8080/create

### List
#### List all buckets owned by the authenticated sender of the request
- http://localhost:8080/list
#### List the files within the specific bucket
- http://localhost:8080/list?bucketname=album-2022-3-20

### Upload
#### Use Uploader from the s3manager package to upload the specific file within the bucket
- http://localhost:8080/upload?bucketname=album-2022-3-20&filename=golang.png
#### Load the files in a local foolder to the specific bucket if the file is not in the bucket yet
- http://localhost:8080/upload?bucketname=album-2022-3-20

### Download 
#### Use Downloader from the s3manager package to download the specific file from the bucket
- http://localhost:8080/download?bucketname=album-2022-3-20&filename=golang.png
#### Look up the specific file from all of the buckets and download it
- http://localhost:8080/download?filename=aws_logo.png
#### Download all of the files within the specific bucket
- http://localhost:8080/download?bucketname=album-2022-3-20

### Delete
#### Delete the specific bucket
- http://localhost:8080/delete?bucketname=album-2022-3-20
#### Look up the specific file from all of the buckets and delete it
- http://localhost:8080/delete?filename=aws_logo.png

## Setup
- GO 1.18
- "github.com/aws/aws-sdk-go/aws"
- "github.com/aws/aws-sdk-go/aws/session"
- "github.com/aws/aws-sdk-go/aws/awserr"
- "github.com/aws/aws-sdk-go/service/s3"
- "github.com/aws/aws-sdk-go/service/s3/s3manager"
- "github.com/gorilla/handlers"

## Execution
- go run main.go handler.go s3service.go
