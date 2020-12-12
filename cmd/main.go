package main

import (
	"fmt"
	"io/ioutil"
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

	s3Svc := s3.New(sess)
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
		strtTscrptJobInpt := transcribeservice.StartTranscriptionJobInput{
			IdentifyLanguage:     aws.Bool(true),
			LanguageCode:         aws.String(transcribeservice.LanguageCodeJaJp),
			LanguageOptions:      []*string{aws.String(transcribeservice.LanguageCodeEnUs)},
			Media:                &transcribeservice.Media{MediaFileUri: &upldOutput.Location},
			OutputBucketName:     aws.String(constant.BucketName),
			OutputKey:            aws.String("output/" + filename),
			TranscriptionJobName: aws.String(constant.TrnscrptJobName),
		}
		strtTscrptJobOtpt, err := trscrbSvc.StartTranscriptionJob(&strtTscrptJobInpt)
		if err != nil {
			log.Fatal("failed to start transcript job:", err)
		}
		getTscrptJobInpt := transcribeservice.GetTranscriptionJobInput{
			TranscriptionJobName: aws.String(*strtTscrptJobOtpt.TranscriptionJob.TranscriptionJobName),
		}
		// job実行結果取得
		for true {
			job, _ := trscrbSvc.GetTranscriptionJob(&getTscrptJobInpt)
			jobStatus := fmt.Sprint(job.TranscriptionJob.TranscriptionJobStatus)

			if transcribeservice.TranscriptionJobStatusCompleted == jobStatus {
				// job実行結果書き出し
				outputPath := filepath.Join(constant.RootDir, constant.OutputDir, filename)
				f, _ := os.Create(outputPath)
				defer f.Close()
				_, err := downloader.Download(f, &s3.GetObjectInput{
					Bucket: aws.String(constant.BucketName),
					Key:    aws.String("output/" + filename),
				})
				log.Printf("failed to download file:%s, err:%+v\n", filename, err)
				break
			}
			if transcribeservice.TranscriptionJobStatusInProgress == jobStatus {
				log.Println("in progress...")
				continue
			}
			if transcribeservice.TranscriptionJobStatusFailed == jobStatus {
				log.Println("failed to transcribe. file:", filename)
				break
			}

			// ちょっとインターバル
			time.Sleep(10 * time.Second)
		}
		// オブジェクトの削除
		_, err = s3Svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(constant.BucketName),
			Key:    aws.String("input/" + filename),
		})
		_, err = s3Svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(constant.BucketName),
			Key:    aws.String("output/" + filename),
		})
		// 処理済みファイルを移動
		bkfile, _ := os.Create(filepath.Join(constant.RootDir, constant.BkDir, filename))
		defer bkfile.Close()
		bkdata, _ := ioutil.ReadFile(path)
		bkfile.Write(bkdata)
	}
}
