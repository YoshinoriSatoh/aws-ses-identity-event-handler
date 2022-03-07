package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"m-int.live/event_handler/kms"
	"m-int.live/event_handler/mail_log"
	"m-int.live/event_handler/notification"
	"m-int.live/event_handler/ses"
)

func HandleLambdaEvent(event events.SNSEvent) error {
	var err error
	fmt.Printf("%#v", event)
	for _, record := range event.Records {

		// Configuration Setで設定されたイベント内容(raw)を保存
		err = mail_log.Logging(record.SNS.Message)
		if err != nil {
			notification.NotifyErrorSlack("Logging mail contents failure")
			return err
		}

		// SESイベント処理
		err = ses.Handler(record.SNS.Message)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var err error
	var logGroupName = os.Getenv("MAIL_LOG_GROUP_NAME")
	var logStreamName = os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")
	err = mail_log.Init(logGroupName, logStreamName)
	if err != nil {
		log.Fatalf("%v", err)
	}

	var channelName = os.Getenv("SLACK_CHANNEL_NAME")
	var slackBotTokenEncrypted = os.Getenv("SLACK_BOT_TOKEN")
	slackBotToken, err := kms.Decrypt(slackBotTokenEncrypted)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = notification.Init(slackBotToken, channelName)
	if err != nil {
		log.Fatalf("%v", err)
	}

	lambda.Start(HandleLambdaEvent)
}
