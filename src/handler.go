package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//Create a S3 bucket named as "album-Year-Month-Date"
func create(w http.ResponseWriter, r *http.Request) {
	year, month, date := time.Now().Date()
	bucketName := "album-" + strconv.Itoa(year) + "-" + strconv.Itoa(int(month)) + "-" + strconv.Itoa(date)
	_, errFlag := createBucket(bucketName)
	if !errFlag {
		w.Write([]byte("Created Bucket " + bucketName))
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

//List of all buckets in AWS
//List the files within the bucket if the bucket name is specific
func list(w http.ResponseWriter, r *http.Request) {
	if bucketName, ok := r.URL.Query()["bucketname"]; ok {
		//List the contents of the bucket
		for _, object := range listObjects(bucketName[0]).Contents {
			fmt.Fprintf(w, *object.Key+"\n")
		}
	} else {
		//List the buckets in the account and region
		allBuckets := getBuckets()
		for _, bkn := range allBuckets {
			fmt.Fprintf(w, bkn+"\n")
		}
	}
}

//Use Uploader from the s3manager package to upload the specific bucket and file
//Upload files from a local folder to the specified bucket if the file is not in the bucket yet
//Parsing URL to get the bucket name and file name
func upload(w http.ResponseWriter, r *http.Request) {
	//Create the session used by S3 Uploader
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(REGION),
	}))

	//Parse URL to get the specific bucket name and file name
	bucketName, okb := r.URL.Query()["bucketname"]
	fileName, okf := r.URL.Query()["filename"]

	//Use uploader to upload the specific file in the bucket
	if okb && okf {
		// Create a downloader with the session
		uploader := s3manager.NewUploader(sess)

		f, err := os.OpenFile(fileName[0], os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return
		}

		// AWS upload manager
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucketName[0]),
			Key:    aws.String(fileName[0]),
			Body:   f,
		})
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Uploaded " + fileName[0] + " " + "to " + aws.StringValue(&result.Location))
		w.Write([]byte("Uploaded " + fileName[0] + " " + "to " + aws.StringValue(&result.Location)))
	} else if okb {
		//Load the contents in local foolder "images" to the specified bucket
		fp := filepath.Join("./", "images")
		if files, err := os.ReadDir(fp); err == nil {
			log.Println("Read Files", files)
			for _, file := range files {
				if !file.IsDir() {
					if flag, err := keyExists(bucketName[0], file.Name()); err == nil && !flag {
						uploadObject(bucketName[0], fp+"/"+file.Name())
						fmt.Fprintf(w, file.Name()+"\n")
					} else {
						if err == nil {
							log.Println("Target file exists in the Bucket")
						} else {
							log.Println(err)
						}
					}
				}
			}
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

//If both bucket and file are specified, use Downloader to download it
//If file is specified, look up the specific file from all of the buckets and download it
//If bucket is specified, download all of the files in this bucket
//Save the dowloaded files in the local folder "download"
func download(w http.ResponseWriter, r *http.Request) {
	//Get all of the buckes in S3
	allBuckets := getBuckets()

	//Create the session used by S3 Downloader
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(REGION),
	}))

	//Parse URL to get the specific bucket name and file name
	bucketName, okb := r.URL.Query()["bucketname"]
	fileName, okf := r.URL.Query()["filename"]

	if okb && okf {
		// Create a downloader with the session
		downloader := s3manager.NewDownloader(sess)

		f, err := os.OpenFile(fileName[0], os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Println(err)
			return
		}

		// Write the contents of S3 Object to the file
		n, err := downloader.Download(f, &s3.GetObjectInput{
			Bucket: aws.String(bucketName[0]),
			Key:    aws.String(fileName[0]),
		})
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Downloaded " + fileName[0] + " " + "Bytes " + fmt.Sprintf("%d", n))
		w.Write([]byte("Downloaded " + fileName[0] + " " + "Bytes " + fmt.Sprintf("%d", n)))
	} else if okb {
		//Download all of the objects in this bucket
		for _, bkn := range allBuckets {
			if bkn == bucketName[0] {
				for _, object := range listObjects(bucketName[0]).Contents {
					s3Resp := getObject(bucketName[0], *object.Key)
					err := writeFile(s3Resp, *object.Key)
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.Write([]byte("Downloaded " + *object.Key + "\n"))
				}
				break
			}
		}
	} else if okf {
		//Look up the specific file from all of the buckets and download it
		for _, bkn := range allBuckets {
			if flag, err := keyExists(bkn, fileName[0]); err == nil && !flag {
				continue
			}

			found := false
			for _, object := range listObjects(bkn).Contents {
				if fileName[0] == *object.Key {
					s3Resp := getObject(bkn, *object.Key)
					err := writeFile(s3Resp, *object.Key)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.Write([]byte("Downloaded " + fileName[0] + "\n"))
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	} else {
		log.Println("Incorrect URL ", r.URL)
		w.WriteHeader(http.StatusBadRequest)
	}
}

//Delete the S3 bucket
//Look up the specific file from all of the buckets and delete it
func delete(w http.ResponseWriter, r *http.Request) {
	allBuckets := getBuckets()

	// All objects in the bucket must be deleted before the bucket itself can be deleted
	if bucketName, ok := r.URL.Query()["bucketname"]; ok {
		//Delete the objects of the specific bucket
		for _, bkn := range allBuckets {
			if bkn == bucketName[0] {
				for _, object := range listObjects(bkn).Contents {
					deleteObject(bkn, *object.Key)
					w.Write([]byte("Deleted " + *object.Key + "\n"))
				}
			}

			//Delete the target bucket
			s3Client.DeleteBucket(&s3.DeleteBucketInput{
				Bucket: aws.String(bucketName[0]),
			})
			break
		}
	} else if fileName, ok := r.URL.Query()["filename"]; ok {
		for _, bkn := range allBuckets {
			if flag, err := keyExists(bkn, fileName[0]); err == nil && !flag {
				continue
			}

			found := false
			for _, object := range listObjects(bkn).Contents {
				if fileName[0] == *object.Key {
					deleteObject(bkn, *object.Key)
					w.Write([]byte("Deleted " + *object.Key + "\n"))
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	} else {
		log.Println("Incorrect URL ", r.URL)
		w.WriteHeader(http.StatusBadRequest)
	}
}
