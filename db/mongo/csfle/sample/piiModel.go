package sample

import (
	"context"
	"os"
	"time"

	"github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/db/mongo/csfle"
	"github.com/sabariramc/goserverbase/v5/log"
)

type Address struct {
	AddressLine1 string `bson:"addressLine1"`
	AddressLine2 string `bson:"addressLine2"`
	AddressLine3 string `bson:"addressLine3"`
	State        string `bson:"state"`
	PIN          string `bson:"pin"`
	Country      string `bson:"country"`
}

type Name struct {
	First  string `bson:"first"`
	Middle string `bson:"middle"`
	Last   string `bson:"last"`
	Full   string `bson:"full"`
}

type PIITestVal struct {
	DOB     time.Time `bson:"dob"`
	Name    Name      `bson:"name"`
	Pan     string    `bson:"pan"`
	Email   string    `bson:"email"`
	Phone   []string  `bson:"phone"`
	Address Address   `bson:"address"`
	UUID    string    `bson:"UUID"`
}

func GetRandomData(uuid string) PIITestVal {
	dob, _ := time.Parse(time.DateOnly, "2001-01-01")
	return PIITestVal{
		UUID: uuid,
		DOB:  dob,
		Name: Name{
			First:  uuid + " first name",
			Middle: uuid + " middle name",
			Last:   uuid + " last name",
			Full:   uuid + " full name",
		},
		Pan:   "ABCDE1234F",
		Email: "abc@" + uuid + ".com",
		Phone: []string{"9600000000", "9600000001"},
		Address: Address{
			AddressLine1: uuid + "address first line",
			AddressLine2: uuid + "address first line",
			AddressLine3: uuid + "address first line",
			State:        "Delhi",
			PIN:          "100000",
			Country:      "India",
		},
	}
}

type LocalMaster struct {
	key string
}

func (a *LocalMaster) Name() string {
	return "local"
}

func (a *LocalMaster) Credentials() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"local": {
			"key": []byte("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		},
	}
}

func (a *LocalMaster) DataKeyOpts() interface{} {
	return nil
}

func GetKMSProvider(ctx context.Context, log log.Log, kmsArn string) (csfle.MasterKeyProvider, error) {
	if os.Getenv("KMS_PROVIDER") == "localstack" {
		cred, _ := aws.GetDefaultAWSConfig().Credentials.Retrieve(ctx)
		kmsProvider := csfle.NewAWSProvider(map[string]interface{}{
			"accessKeyId":     cred.AccessKeyID,
			"secretAccessKey": cred.SecretAccessKey,
		}, csfle.AWSDataKeyOpts{
			Region:   aws.GetDefaultAWSConfig().Region,
			KeyARN:   kmsArn,
			Endpoint: "http://localhost.localstack.cloud:4566",
		})
		return kmsProvider, nil
	} else if os.Getenv("KMS_PROVIDER") == "local" {
		return &LocalMaster{}, nil
	}
	return csfle.GetAWSMasterKeyProvider(ctx, log, kmsArn)
}
