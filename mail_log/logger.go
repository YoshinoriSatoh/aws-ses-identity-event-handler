package mail_log

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

var client = cloudwatchlogs.New(session.Must(session.NewSession()))
var logGroupName string
var logStreamName string
var sequenceToken *string

func Init(setLogGroupName, setLogStreamName string) error {
	logGroupName = setLogGroupName
	logStreamName = setLogStreamName
	logDescription, err := client.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, stream := range logDescription.LogStreams {
		if *stream.LogStreamName == logStreamName {
			if stream.UploadSequenceToken == nil {
				sequenceToken = nil
			} else {
				sequenceToken = stream.UploadSequenceToken
			}
			return nil
		}
	}
	_, err = client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})
	if err != nil {
		return err
	}
	sequenceToken = nil
	return nil
}

func Logging(message string) error {
	putLogEventsOutput, err := client.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			{
				Message:   aws.String(message),
				Timestamp: aws.Int64((time.Now().UnixMilli())),
			},
		},
		SequenceToken: sequenceToken,
	})
	sequenceToken = putLogEventsOutput.NextSequenceToken
	if err != nil {
		return err
	}
	return nil
}
