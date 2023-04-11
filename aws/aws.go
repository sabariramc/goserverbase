package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go/aws"
)

var defaultAWSSession *session.Session

func SetDefaultAWSSession(defaultSession *session.Session) {
	defaultAWSSession = awstrace.WrapSession(defaultSession)
}

func GetDefaultAWSSession() *session.Session {
	return defaultAWSSession
}

func NewRegionalDefaultAWSSession(region string) *session.Session {
	return NewRegionalAWSSession(defaultAWSSession, region)
}

func NewRegionalAWSSession(awsSession *session.Session, region string) *session.Session {
	newSession := session.Must(session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: awsSession.Config.Credentials,
	}))
	return awstrace.WrapSession(newSession)
}
