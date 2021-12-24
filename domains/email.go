package domains

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/smtp"
	"os"
	"time"
)

type Email struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	To      string             `bson:"to" json:"to"`
	Subject string             `bson:"subject" json:"subject"`
	Message string             `bson:"message" json:"message"`
	Created time.Time          `bson:"created" json:"created"`
}

func GetAllEmails(db *mongo.Database) ([]Email, error) {
	col := db.Collection("Emails")
	cur, err := col.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var contents []Email
	if err := cur.All(context.TODO(), &contents); err != nil {
		return nil, err
	}

	return contents, nil
}

func CreateEmail(db *mongo.Database, c Email) (*Email, error) {
	err := SendEmail(c.To, c.Subject, c.Message)
	if err != nil {
		return nil, err
	}
	col := db.Collection("Emails")
	c.Created = time.Now()
	r, err := col.InsertOne(context.TODO(), c)
	if err != nil {
		return nil, err
	}

	c.ID = r.InsertedID.(primitive.ObjectID)
	return &c, nil
}

// SendEmail sends email
func SendEmail(to, subject, message string) error {
	username := os.Getenv("SMTP_username")
	password := os.Getenv("SMTP_password")
	host := os.Getenv("SMTP_host")
	port := os.Getenv("SMTP_port")
	from := os.Getenv("SMTP_from")

	auth := smtp.PlainAuth("", username, password, host)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte("To: " + to + "\r\n" + "From: " + from + "\r\n" + "Subject: " + subject + "\r\n" + mime + message)
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, msg)

	return err
}

// SendEmailFromAWS not used currently. But can be used to send emails directly from aws server
func SendEmailFromAWS(to, subject, message string) {
	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)

	charSet := "UTF-8"
	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{
			},
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(message),
				},
				/*Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody),
				},*/
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String("test@example.com"),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}

		return
	}

	log.Println("Email Sent to address: " + to)
	log.Println(result)
}
