package s3

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func S3Client() {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"2A4Q92QCM7NC2387KQJL",
			"JYnxueanmgtrMyvRxfUTu7qXUveEvQOJliomlsIz",
			"us-east-1",
		)),
		config.WithBaseEndpoint("https://172.38.30.192:8480"),
		config.WithHTTPClient(&http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}),
		//config.WithClientLogMode(aws.LogRequest|aws.LogResponse),
	)

	if err != nil {
		log.Fatal(err)
	}

	// 创建S3客户端 - 保持路径风格设置
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		// https://github.com/aws/aws-sdk-go-v2/blob/release-2025-01-22/service/s3/CHANGELOG.md#v1730-2025-01-15
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		// https://github.com/aws/aws-sdk-go-v2/issues/2974
		o.DisableLogOutputChecksumValidationSkipped = true
	})

	// // 执行存储桶创建操作
	// _, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
	// 	Bucket: aws.String("test-bucket"),
	// })
	// if err != nil {
	// 	log.Fatalf("CreateBucket失败: %v", err)
	// }
	// fmt.Println("存储桶创建成功！")

	// // 执行获取存储桶元数据操作
	// _, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
	// 	Bucket: aws.String("test-bucket"),
	// })
	// if err != nil {
	// 	log.Fatalf("HeadBucket失败: %v", err)
	// }
	// fmt.Println("存储桶元数据获取成功！")

	// 执行上传对象操作

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test.txt"),
		Body:   strings.NewReader("Hello World"),
		//Body:   bytes.NewReader(GenerateRandomBytes(10)),
		//ChecksumAlgorithm: types.ChecksumAlgorithmCrc32,
		// 添加内容类型
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		log.Fatalf("PutObject失败: %v", err)
	}
	fmt.Println("对象上传成功！")

	//// 执行获取对象操作
	//outputGet, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
	//	Bucket: aws.String("test-bucket"),
	//	Key:    aws.String("test.txt"),
	//	//ChecksumMode: types.ChecksumModeEnabled,
	//})
	//if err != nil {
	//	log.Fatalf("GetObject失败: %v", err)
	//}
	//defer outputGet.Body.Close()
	//
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(outputGet.Body)
	//if err != nil {
	//	log.Fatalf("读取对象失败: %v", err)
	//}
	//fmt.Println("获取的对象内容:", buf.String())
}
