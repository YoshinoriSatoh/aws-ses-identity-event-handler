package ses

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"m-int.live/event_handler/notification"
)

var client = sesv2.New(session.Must(session.NewSession()))

// 以下ページのSNS通知コンテンツから、通知判断等に必要な項目のみ構造体として抜粋
// (ログにはMessageフィールドのJSON文字列を全て対応する各種Cloudwatch log groupに保存)
// https://docs.aws.amazon.com/ja_jp/ses/latest/dg/notification-contents.html
type BounceRecipients struct {
	EmailAddress string `json:"emailAddress"`
}

type Bounce struct {
	BounceType        string `json:"bounceType"`
	BounceSubType     string `json:"bounceSubType"`
	BouncedRecipients []BounceRecipients
	Timestamp         string `json:"timestamp"`
}

type ComplainedRecipient struct {
	EmailAddress string `json:"emailAddress"`
}

type Complaint struct {
	ComplainedRecipients  []ComplainedRecipient
	ComplaintFeedbackType string `json:"complaintFeedbackType"`
	Timestamp             string `json:"timestamp"`
}

type Message struct {
	EventType string    `json:"eventType"`
	Bounce    Bounce    `json:"bounce"`
	Complaint Complaint `json:"complaint"`
}

// 以下の通知種別に応じて、Slackチャネル通知及びサプレッションリストへの追加を実施
// https://docs.aws.amazon.com/ja_jp/ses/latest/dg/notification-contents.html
func Handler(messageJson string) error {
	var message Message

	err := json.Unmarshal([]byte(messageJson), &message)
	if err != nil {
		fmt.Println(err)
		return err
	}

	switch message.EventType {
	case "Bounce":
		err = bounceHandler(message)
	case "Complaint":
		err = complaintHandler(message)
	}
	if err != nil {
		return err
	}
	return nil
}

// 即座にサプレッションリストに登録するのは Permanent の General, NoEmail のみ
// それ以外は通知等を確認の上、必要に応じて手動でサプレッションリストに登録する
func bounceHandler(message Message) error {
	var err error
	notificationMessage := fmt.Sprintf("Type=%s, SubType=%s", message.Bounce.BounceType, message.Bounce.BounceSubType)
	reason := "BOUNCE"

	switch message.Bounce.BounceType {
	case "Undetermined":
		switch message.Bounce.BounceSubType {
		case "Undetermined":
			for _, r := range message.Bounce.BouncedRecipients {
				notification.NotifyEventSlack(message.EventType, notificationMessage, r.EmailAddress)
			}
		}

	case "Permanent":
		switch message.Bounce.BounceSubType {
		case "General":
			fallthrough
		case "NoEmail":
			for _, r := range message.Bounce.BouncedRecipients {
				err = addToSuppressionList(message.EventType, r.EmailAddress, reason)
				notification.NotifyEventSlack(message.EventType, notificationMessage, r.EmailAddress)
			}
		case "Suppressed":
			fallthrough
		case "OnAccountSuppressionList":
			for _, r := range message.Bounce.BouncedRecipients {
				notification.NotifyEventSlack(message.EventType, notificationMessage, r.EmailAddress)
			}
		}
	case "Transient":
		switch message.Bounce.BounceSubType {
		case "General":
			fallthrough
		case "MailboxFull":
			fallthrough
		case "MessageTooLarge":
			fallthrough
		case "ContentRejected":
			fallthrough
		case "AttachmentRejected":
			for _, r := range message.Bounce.BouncedRecipients {
				notification.NotifyEventSlack(message.EventType, notificationMessage, r.EmailAddress)
			}
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// Complaintは通知のみ
// 通知等を確認の上、必要に応じて手動でサプレッションリストに登録する
func complaintHandler(message Message) error {
	notificationMessage := fmt.Sprintf("FeedbackType=%s", message.Complaint.ComplaintFeedbackType)

	switch message.Complaint.ComplaintFeedbackType {
	case "abuse":
		fallthrough
	case "auth-failure":
		fallthrough
	case "fraud":
		fallthrough
	case "not-spam":
		fallthrough
	case "other":
		fallthrough
	case "virus":
		for _, r := range message.Complaint.ComplainedRecipients {
			notification.NotifyEventSlack(message.EventType, notificationMessage, r.EmailAddress)
		}
	}
	return nil
}

func addToSuppressionList(eventType, emailAddress, reason string) error {
	_, err := client.PutSuppressedDestination(&sesv2.PutSuppressedDestinationInput{
		EmailAddress: aws.String(emailAddress),
		Reason:       aws.String(reason),
	})
	if err != nil {
		notificationMessage := fmt.Sprintf("PutSuppressedDestination failure: Message=%s", err.Error())
		notification.NotifyErrorSlack(notificationMessage)
		return err
	}
	return nil
}
