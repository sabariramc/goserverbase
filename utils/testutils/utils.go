package testutils

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	base "github.com/sabariramc/goserverbase/v3/aws"

	"github.com/joho/godotenv"
)

func setAWSSession() {
	// stsToken := getSTSToken()
	// awsConfig := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))
	defaultConfig, _ := config.LoadDefaultConfig(context.TODO())
	// awsSession := session.Must(session.NewSession(&aws.Config{
	// 	Region:      aws.String(os.Getenv("region")),
	// 	Credentials: credentials.NewStaticCredentials(stsToken["AccessKeyId"], stsToken["SecretAccessKey"], stsToken["SessionToken"]),
	// }))
	base.SetDefaultAWSConfig(defaultConfig)
}

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		fmt.Printf("Env file not found - %v\n", path)
	}
}

func Initialize() {
	setAWSSession()
}
