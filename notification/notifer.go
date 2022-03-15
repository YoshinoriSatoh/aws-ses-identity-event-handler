package notification

import (
	"fmt"

	"github.com/slack-go/slack"
)

var client *slack.Client
var channelName string
var envName string

func Init(slackBotToken, initChannelName, initEnvName string) error {

	client = slack.New(slackBotToken)
	channelName = initChannelName
	envName = initEnvName
	return nil
}

func NotifyErrorSlack(message string) error {
	_, _, err := client.PostMessage(channelName, slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*SES EventHander Error Notification*\n %s", message),
			},
			nil,
			nil,
		),
	))
	if err != nil {
		fmt.Printf("slack error: %#v\n", err)
		return err
	}
	return nil
}

func NotifyEventSlack(eventType, message, emailAddress string) error {
	_, _, err := client.PostMessage(channelName, slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*SES Event Notification*\n  Environment: %s\n  EventType: %s\n  EmailAddress: %s\n  Message: %s",
					envName, eventType, emailAddress, message),
			},
			nil,
			nil,
		),
	))
	if err != nil {
		fmt.Printf("slack error: %#v\n", err)
		return err
	}
	return nil
}
