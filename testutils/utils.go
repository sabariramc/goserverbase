package testutils

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	base "github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/instrumentation"
	"github.com/sabariramc/goserverbase/v5/utils"

	"github.com/joho/godotenv"
)

func SetAWSSession(tr instrumentation.Tracer) {
	cnf, _ := config.LoadDefaultConfig(context.TODO())
	if utils.GetEnv("AWS_PROVIDER", "") == "local" {
		var err error
		cnf, err = base.GetLocalStackConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	base.SetDefaultAWSConfig(cnf, tr)
}

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		fmt.Printf("Env file error - %v\n", err)
	}
}

func Initialize() {
	SetAWSSession(nil)
}
