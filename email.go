package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"log"
	"text/template"
)

var (
	//go:embed "templates/updateEmail.txt"
	updateEmailText string
)

const (
	SourceEmail = "a4web@arran.net.au"
)

func notifyChange(ctx context.Context, svc sesiface.SESAPI, email string, page string) error {
	// TODO Make this periodically check db to see if there is anything queued, and normally queue up messages
	if email == "" {
		return fmt.Errorf("no email specified")
	}

	if svc == nil {
		return fmt.Errorf("no email provider specified")
	}
	from := SourceEmail

	type EmailContent struct {
		To      string
		From    string
		Subject string
		URL     string
	}

	// Define email content
	content := EmailContent{
		To:      email,
		From:    from,
		Subject: "Website Update Notification",
		URL:     page,
	}

	// Create a new buffer to store the rendered email content
	var notification bytes.Buffer

	// Parse and execute the email template
	tmpl, err := template.New("email").Parse(updateEmailText)
	if err != nil {
		return fmt.Errorf("parse email template: %w", err)
	}

	// Execute the template and store the result in the notification buffer
	err = tmpl.Execute(&notification, content)
	if err != nil {
		return fmt.Errorf("execute email template: %w", err)
	}

	// Compose the email
	destination := &ses.Destination{
		ToAddresses: []*string{aws.String(email)},
	}

	message := &ses.Message{
		Subject: &ses.Content{
			Charset: aws.String("UTF-8"),
			Data:    aws.String(content.Subject),
		},
		Body: &ses.Body{
			Text: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(notification.String()),
			},
		},
	}

	input := &ses.SendEmailInput{
		Destination: destination,
		Message:     message,
		Source:      aws.String(from),
	}

	// Send email using AWS SES
	result, err := svc.SendEmailWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	log.Println("Email sent successfully:", *result.MessageId)
	return nil
}

func getEmailProvider() sesiface.SESAPI {
	// TODO
	return nil
}
