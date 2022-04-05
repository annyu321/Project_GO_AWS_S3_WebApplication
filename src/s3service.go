package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

//Create a new bucket
func createBucket(bucketName string) (resp *s3.CreateBucketOutput, errFlag bool) {
	errFlag = false
	resp, err := s3Client.CreateBucket(&s3.CreateBucketInput{
		// ACL: aws.String(s3.BucketCannedACLPrivate),
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(REGION),
		},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				log.Println("Bucket name already in use!")
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				log.Println("Bucket exists and is owned by you!")
			default:
				errFlag = true
			}
		}
	}
	return resp, errFlag
}

//Add an object to a bucket.
func uploadObject(bucketName, filename string) (resp *s3.PutObjectOutput) {
	f, err := os.Open(filename)
	if err != nil {
		log.Println(err)
	}

	log.Println("Upload ", filename)
	resp, err = s3Client.PutObject(&s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(bucketName),
		Key:    aws.String(strings.Split(filename, "/")[1]),
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
	})

	if err != nil {
		log.Println(err)
	}
	return resp
}

//List objects within the bucket
func listObjects(bucketName string) (resp *s3.ListObjectsV2Output) {
	resp, err := s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		log.Println(err)
	}
	return resp
}

//Retrieve objects from S3 with the specific bucket and file
func getObject(bucketName, fileName string) (resp *s3.GetObjectOutput) {
	log.Println("Download ", fileName)
	resp, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	if err != nil {
		log.Println(err)
	}
	return resp
}

//Delete the object
func deleteObject(bucketName, fileName string) (resp *s3.DeleteObjectOutput) {
	log.Println("Delete ", fileName)
	resp, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	if err != nil {
		log.Println(err)
	}
	return resp
}

//List all buckets owned by the authenticated sender of the request
func getBuckets() []string {
	bkNames := []string{}
	result, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		return nil
	}

	for _, bk := range result.Buckets {
		bkNames = append(bkNames, *bk.Name)
	}
	log.Println(bkNames)
	return bkNames
}

//Retrieve the objects to check if it exists in the bucket
func keyExists(bucket, key string) (bool, error) {
	_, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

//Write the s3 response body to the specific file
func writeFile(s3resp *s3.GetObjectOutput, fileName string) error {
	fp := filepath.Join("./", "download")
	if body, err := ioutil.ReadAll(s3resp.Body); err == nil {
		if werr := os.WriteFile(fp+"/"+fileName, body, 0644); werr != nil {
			log.Println(werr)
		}
	} else {
		log.Println(err)
		return err
	}
	return nil
}
