package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/transcribeservice"
	"github.com/fujimaru-lab/transcribe_go/internal/constant"
)

func main() {
	// sessionの作成
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(constant.Region),
	})
	if err != nil {
		log.Fatal("failed to get session:", err)
	}

	// s3Svc := s3.New(sess)
	// 未アップロードの音声ファイルをリストアップ
	var inptFilePaths []string
	err = filepath.Walk(filepath.Join(constant.RootDir, constant.InputDir),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Println("Failed to access the filepath ->", path, ":", err)
				return err
			} else if info.IsDir() || info.Name() == ".gitkeep" {
				return nil
			} else {
				inptFilePaths = append(inptFilePaths, path)
				return nil
			}
		})
	if err != nil {
		log.Fatal("failed to list input files")
	}
	if len(inptFilePaths) == 0 {
		log.Println("No input file exists")
		return
	}

	// 音声ファイルアップロード
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	trscrbSvc := transcribeservice.New(sess)
	for _, path := range inptFilePaths {
		filename := filepath.Base(path)
		log.Printf("start to upload file: %s", filename)
		f, _ := os.Open(path)
		defer f.Close()
		upldInput := &s3manager.UploadInput{
			Bucket: aws.String(constant.BucketName),
			Key:    aws.String("input/" + filename),
			Body:   f,
		}
		upldOutput, nil := uploader.Upload(upldInput)
		if err != nil {
			log.Println("failed to upload file:", filename)
		}

		// Transcribe job実行
		now := time.Now()
		datetimeSuffix := fmt.Sprintf("%d%d_%d%d%d", now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
		strtTscrptJobInpt := transcribeservice.StartTranscriptionJobInput{
			LanguageCode:         aws.String(transcribeservice.LanguageCodeJaJp),
			Media:                &transcribeservice.Media{MediaFileUri: &upldOutput.Location},
			OutputBucketName:     aws.String(constant.BucketName),
			OutputKey:            aws.String("output/" + filename + ".json"),
			TranscriptionJobName: aws.String(fmt.Sprintf("%s%s", constant.TrnscrptJobName, datetimeSuffix)),
		}
		log.Printf("start to transcribe. param: %v\n", strtTscrptJobInpt)
		strtTscrptJobOtpt, err := trscrbSvc.StartTranscriptionJob(&strtTscrptJobInpt)
		if err != nil {
			log.Fatal("failed to start transcript job:", err)
		}
		getTscrptJobInpt := transcribeservice.GetTranscriptionJobInput{
			TranscriptionJobName: aws.String(*strtTscrptJobOtpt.TranscriptionJob.TranscriptionJobName),
		}
		// job実行結果取得
		isNotFin := true
		for isNotFin {
			job, _ := trscrbSvc.GetTranscriptionJob(&getTscrptJobInpt)
			jobStatus := *job.TranscriptionJob.TranscriptionJobStatus

			if transcribeservice.TranscriptionJobStatusCompleted == jobStatus {
				log.Printf("job is %s", jobStatus)
				// job実行結果書き出し
				tmp := filepath.Join(constant.RootDir, constant.OutputDir, filename)
				outputPath := tmp + ".json"
				f, err := os.Create(outputPath)
				if err != nil {
					log.Printf("failed to create download file")
				}
				defer f.Close()
				key := "output/" + filename + ".json"
				_, err = downloader.Download(f, &s3.GetObjectInput{
					Bucket: aws.String(constant.BucketName),
					Key:    aws.String(key),
				})
				if err != nil {
					log.Printf("failed to download Key:%s, err:%+v\n", key, err)
					log.Printf("CHECK OUT S3: %s\n", *job.TranscriptionJob.Transcript.TranscriptFileUri)
				} else {
					log.Printf("succeed at download file:%s\n", filename)
				}
				isNotFin = false
			}
			if transcribeservice.TranscriptionJobStatusInProgress == jobStatus {
				log.Printf("job is %s......", jobStatus)
				isNotFin = true
			}
			if transcribeservice.TranscriptionJobStatusFailed == jobStatus {
				log.Printf("job is %s. file:%s", jobStatus, filename)
				isNotFin = false
			}
			// ちょっとインターバル
			time.Sleep(5 * time.Second)
		}

	}
}
