package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/handlers"
)

var s3Client *s3.S3

const REGION = "ap-northeast-1"

func init() {
	//Create S3 service client
	s3Client = s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(REGION),
	})))
}

func newLoggingHandler(dst io.Writer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(dst, h)
	}
}

func main() {
	terminate := make(chan os.Signal)
	mux := http.NewServeMux()

	//Use the LoggingHandler middleware from the gorilla/handlers package
	logFile, err := os.OpenFile("server.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	
	loggingHandler := newLoggingHandler(logFile)

	//Create a new S3 bucket named as "album-Year-Month-Date"
	mux.Handle("/create", loggingHandler(http.HandlerFunc(create)))

	//List of all buckets in AWS
	//List the files within the bucket if the bucket name is specific
	mux.Handle("/list", loggingHandler(http.HandlerFunc(list)))

	//Use Uploader from the s3manager package to upload the specific file in the bucket
	//Load the files in a local foolder to the specific bucket if the file is not in the bucket yet
	mux.Handle("/upload", loggingHandler(http.HandlerFunc(upload)))

	//use Downloader from the s3manager package to download the specific file from the S3 bucket
	//Look up the specific file from all of the buckets and download it
	//Download all of files in the specific bucket
	mux.Handle("/download", loggingHandler(http.HandlerFunc(download)))

	//Delete the S3 bucket
	//Delete the file if the bucket and file are specified
	mux.Handle("/delete", loggingHandler(http.HandlerFunc(delete)))

	log.Println("Start Http Server")
	go func() {
		err := http.ListenAndServe(":8080", mux)
		log.Printf("Stopped HTTP server: %s", err)
	}()

	<-terminate
}
