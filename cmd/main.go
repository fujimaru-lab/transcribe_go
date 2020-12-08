package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fujimaru-lab/transcribe_go/internal/constraint"
)

func main() {
	// sessionの作成
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(constraint.Region),
	})
	if err != nil {
		log.Fatal("failed to get session:", err)
		return
	}

	// S3バケット作成
	s3Svc := s3.New(sess)
	bucket, err := createBucket(s3Svc)
	if err != nil {
		log.Fatal("failed to create bucket:", err)
		return
	}

	// 未アップロードの音声ファイルをリストアップ

	// 音声ファイルアップロード

	// Transcribe job実行

	// job実行結果取得

	// job実行結果書き出し

	// 処理済みファイル更新

}

func createBucket(s3Svc *s3.S3) (*s3.CreateBucketOutput, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(constraint.BucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(constraint.Region),
		},
	}
	output, err := s3Svc.CreateBucket(input)
	return output, err
}
