package testutils

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	base "github.com/sabariramc/goserverbase/v4/aws"

	"github.com/joho/godotenv"
)

func setAWSSession() {
	defaultConfig, _ := config.LoadDefaultConfig(context.TODO())
	base.SetDefaultAWSConfig(defaultConfig)
}

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		fmt.Printf("Env file error - %v\n", err)
	}
}

func Initialize() {
	setAWSSession()
}
