package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var s3Client *s3.S3

const REGION = "ap-northeast-1"

func init() {
	//Create S3 service client
	s3Client = s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(REGION),
	})))
}

func main() {
	terminate := make(chan os.Signal)
	mux := http.NewServeMux()

	//Create a new S3 bucket named as "album-Year-Month-Date"
	mux.HandleFunc("/create", create)

	//List of all buckets in AWS
	//List the files within the bucket if the bucket name is specific
	mux.HandleFunc("/list", list)

	//Use Uploader from the s3manager package to upload the specific file in the bucket
	//Load the files in a local foolder to the specific bucket if the file is not in the bucket yet
	mux.HandleFunc("/upload", upload)

	//use Downloader from the s3manager package to download the specific file from the S3 bucket
	//Look up the specific file from all of the buckets and download it
	//Download all of files in the specific bucket
	mux.HandleFunc("/download", download)

	//Delete the S3 bucket
	//Delete the file if the bucket and file are specified
	mux.HandleFunc("/delete", delete)

	log.Println("Start Http Server")
	go func() {
		err := http.ListenAndServe(":8080", mux)
		log.Printf("Stopped HTTP server: %s", err)
	}()

	<-terminate
}
