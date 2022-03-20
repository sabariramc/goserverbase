package testutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	base "sabariram.com/goserverbase/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/joho/godotenv"
)

func getSTSToken() map[string]string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	filename := fmt.Sprintf("%v/.aws/sts.json", dirname)
	rawData, _ := ioutil.ReadFile(filename)
	var envData map[string]interface{}
	err = json.Unmarshal(rawData, &envData)
	if err != nil {
		panic(err)
	}
	rawCredentials := envData["Credentials"].(map[string]interface{})
	credentials := make(map[string]string)
	for key, value := range rawCredentials {
		credentials[key] = value.(string)
	}
	return credentials
}

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
		fmt.Printf("Env file not found - %v", path)
	}
}

func Initialize() {
	// getSTSToken()
	setAWSSession()

}
