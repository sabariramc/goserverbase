package testutils

import (
	"fmt"

	base "github.com/sabariramc/goserverbase/v3/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/joho/godotenv"
)

func setAWSSession() {
	// stsToken := getSTSToken()
	// awsSession := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))
	awsSession := session.Must(session.NewSession())
	// awsSession := session.Must(session.NewSession(&aws.Config{
	// 	Region:      aws.String(os.Getenv("region")),
	// 	Credentials: credentials.NewStaticCredentials(stsToken["AccessKeyId"], stsToken["SecretAccessKey"], stsToken["SessionToken"]),
	// }))
	base.SetDefaultAWSSession(awsSession)
}

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		fmt.Printf("Env file not found - %v\n", path)
	}
}

func Initialize() {
	setAWSSession()
}
