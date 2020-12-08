package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fujimaru-lab/transcribe_go/internal/constant"
	"github.com/fujimaru-lab/transcribe_go/internal/targetfile"
)

func main() {
	// sessionの作成
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(constant.Region),
	})
	if err != nil {
		log.Fatal("failed to get session:", err)
	}

	// S3バケット作成
	s3Svc := s3.New(sess)
	crtBktInput := &s3.CreateBucketInput{
		Bucket: aws.String(constant.BucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(constant.Region),
		},
	}
	crtBktOutput, err := s3Svc.CreateBucket(crtBktInput)
	if err != nil {
		log.Fatal("failed to create bucket:", err)
	}

	// 未アップロードの音声ファイルをリストアップ
	filepaths, err := targetfile.ListTargetFilePath()
	if err != nil {
		log.Fatal("failed to list Transcribed file:", err)
	}

	// 音声ファイルアップロード

	// Transcribe job実行

	// job実行結果取得

	// job実行結果書き出し

	// 処理済みファイル更新

}
