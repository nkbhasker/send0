package service

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/model"
)

type snsNotificationPayload struct {
	Type             string  `json:"Type"`
	MessageId        string  `json:"MessageId"`
	TopicArn         string  `json:"TopicArn"`
	Message          string  `json:"Message"`
	Signature        string  `json:"Signature"`
	SignatureVersion string  `json:"SignatureVersion"`
	Timestamp        string  `json:"Timestamp"`
	SigingCertURL    string  `json:"SigningCertURL"`
	Subject          *string `json:"Subject"`
	Token            *string `json:"Token"`
	SubscribeURL     *string `json:"SubscribeURL"`
	UnsubscribeURL   *string `json:"UnsubscribeURL"`
}

type sesNotificationMessage struct {
	EventType     types.EventType       `json:"eventType"`
	Mail          mailPayload           `json:"mail"`
	Bounce        *bouncePayload        `json:"bounce"`
	Complaint     *complaintPayload     `json:"complaint"`
	Delivery      *deliveryPayload      `json:"delivery"`
	Reject        *rejectPayload        `json:"reject"`
	Open          *openPayload          `json:"open"`
	DeliveryDelay *deliveryDelayPayload `json:"deliveryDelay"`
}

type mailPayload struct {
	MessageId        string   `json:"messageId"`
	Source           string   `json:"source"`
	Timestamp        string   `json:"timestamp"`
	Destination      []string `json:"destination"`
	HeadersTruncated bool     `json:"headersTruncated"`
	CommonHeaders    struct {
		From []string `json:"from"`
		To   []string `json:"to"`
	} `json:"commonHeaders"`
}

type bouncePayload struct {
	BounceType        string `json:"bounceType"`
	BouncedRecipients []struct {
		EmailAddress string `json:"emailAddress"`
		Status       string `json:"status"`
		Action       string `json:"action"`
	} `json:"bouncedRecipients"`
	Timestamp string `json:"timestamp"`
}

type complaintPayload struct {
	ComplaintFeedbackType string `json:"complaintFeedbackType"`
	ComplainedRecipients  []struct {
		EmailAddress string `json:"emailAddress"`
	} `json:"complainedRecipients"`
	Timestamp string `json:"timestamp"`
}

type deliveryPayload struct {
	ProcessingTimeMillis int `json:"processingTimeMillis"`
	Recipients           []struct {
		EmailAddress string `json:"emailAddress"`
	} `json:"recipients"`
	Timestamp string `json:"timestamp"`
}

type rejectPayload struct {
	Reason string `json:"reason"`
}

type openPayload struct {
	Timestamp string `json:"timestamp"`
	IpAddress string `json:"ipAddress"`
	UserAgent string `json:"userAgent"`
}

type deliveryDelayPayload struct {
	DelayType         string `json:"delayType"`
	ExpirationTime    string `json:"expirationTime"`
	DelayedRecipients []struct {
		EmailAddress string `json:"emailAddress"`
		Status       string `json:"status"`
	} `json:"delayedRecipients"`
	Timestamp string `json:"timestamp"`
}

type snsService struct {
	*baseService
	mu                  sync.RWMutex
	eventService        EventSevice
	svcs                map[constant.AwsRegion]*sns.Client
	snsTopicArns        map[constant.AwsRegion]*model.SNSTopic
	signingCertificates map[string]*x509.Certificate
}

type SNSService interface {
	Topics() map[constant.AwsRegion]*model.SNSTopic
	SetupTopics(ctx context.Context) error
	Subscribe(ctx context.Context, region constant.AwsRegion, topicArn string) error
	ConfirmSubscribe(payload []byte) error
	ConfirmUnsubscribe(payload []byte) error
	ProcessNotification(payload []byte) error
}

func NewSNSService(baseService *baseService, eventService EventSevice) (SNSService, error) {
	svcs := make(map[constant.AwsRegion]*sns.Client)
	for _, region := range constant.SupportedSESRegions {
		svcs[region] = sns.New(sns.Options{
			Region: string(region),
			Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
				baseService.config.SES.AccessKeyId,
				baseService.config.SES.SecretAccessKey,
				"",
			)),
		})
	}
	return &snsService{
		baseService:         baseService,
		eventService:        eventService,
		svcs:                svcs,
		snsTopicArns:        make(map[constant.AwsRegion]*model.SNSTopic),
		signingCertificates: make(map[string]*x509.Certificate),
	}, nil
}

func (s *snsService) Topics() map[constant.AwsRegion]*model.SNSTopic {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snsTopicArns
}

func (s *snsService) SetupTopics(ctx context.Context) error {
	s.logger.Info().Msg("Setting up SNS topics")
	topics, err := s.repository.SNSTopic.FindAll(context.Background())
	if err != nil {
		return err
	}
	snsTopicArns := make(map[constant.AwsRegion]*model.SNSTopic)
	for region, svc := range s.svcs {
		topicIdx := slices.IndexFunc(topics, func(topic *model.SNSTopic) bool {
			return topic.Region == region
		})
		if topicIdx >= 0 {
			s.logger.Info().Str("region", string(region)).Msg("Found topic")
			snsTopicArns[region] = topics[topicIdx]
			continue
		}
		s.logger.Info().Str("region", string(region)).Msg("Creating new topic")
		resp, err := svc.CreateTopic(ctx, &sns.CreateTopicInput{
			Name: aws.String("ses" + "-" + constant.AppName),
			Attributes: map[string]string{
				"FifoTopic": "false",
			},
		})
		if err != nil {
			return err
		}
		newTopic := &model.SNSTopic{
			Region: region,
			Arn:    *resp.TopicArn,
			Status: constant.AwsSNSTopicStatusPending,
		}
		err = s.repository.SNSTopic.Save(context.Background(), newTopic)
		if err != nil {
			return err
		}
		snsTopicArns[region] = newTopic
	}
	s.mu.Lock()
	s.snsTopicArns = snsTopicArns
	s.mu.Unlock()
	// subscribe to topics
	for region, topic := range snsTopicArns {
		if topic.Status == constant.AwsSNSTopicStatusPending {
			err = s.Subscribe(ctx, region, topic.Arn)
			if err != nil {
				return err
			}
		}
	}
	s.logger.Info().Msg("SNS topics setup complete")

	return nil
}

func (s *snsService) Subscribe(ctx context.Context, region constant.AwsRegion, topicArn string) error {
	s.logger.Info().Str("region", string(region)).Msg("Subscribing to topic")
	_, err := s.svcs[region].Subscribe(ctx, &sns.SubscribeInput{
		Protocol: aws.String("https"),
		TopicArn: aws.String(topicArn),
		Endpoint: aws.String(s.config.SNS.EndPoint),
	})
	if err != nil {
		return err
	}

	return nil

}

func (s *snsService) ConfirmSubscribe(payload []byte) error {
	s.logger.Info().Msg("Confirming subscription")
	region, snsNotification, err := s.parsePayload(payload)
	if err != nil {
		return err
	}
	err = s.verifySignature(*snsNotification)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error verifying signature")
		return err
	}
	s.logger.Info().Str("region", string(*region)).Msg("Starting subscription confirmation")
	resp, err := http.Get(*snsNotification.SubscribeURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("non 200 response on subscription URL")
	}
	s.logger.Info().Str("region", string(*region)).Msg("Subscription confirmed, updating status")
	err = s.repository.SNSTopic.UpdateStatus(context.Background(), *region, constant.AwsSNSTopicStatusActive)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error updating topic status")
		return err
	}
	s.mu.Lock()
	s.snsTopicArns[*region].Status = constant.AwsSNSTopicStatusActive
	s.mu.Unlock()
	s.logger.Info().Str("region", string(*region)).Msg("Topic added to map")

	return nil
}

func (s *snsService) ConfirmUnsubscribe(payload []byte) error {
	region, snsNotification, err := s.parsePayload(payload)
	if err != nil {
		return err
	}
	err = s.verifySignature(*snsNotification)
	if err != nil {
		return err
	}
	resp, err := http.Get(*snsNotification.UnsubscribeURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("non 200 response on unsubscription URL")
	}
	err = s.repository.SNSTopic.UpdateStatus(context.Background(), *region, constant.AwsSNSTopicStatusInactive)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error updating topic status")
		return err
	}
	s.mu.Lock()
	delete(s.snsTopicArns, *region)
	s.mu.Unlock()

	return nil
}

func (s *snsService) ProcessNotification(payload []byte) error {
	_, snsNotification, err := s.parsePayload(payload)
	if err != nil {
		return err
	}
	err = s.verifySignature(*snsNotification)
	if err != nil {
		return err
	}
	if snsNotification.Message == constant.AwsSESEventTopicSuccessMessage {
		s.logger.Info().Msg(snsNotification.Message)
		return nil
	}
	var message sesNotificationMessage
	err = json.Unmarshal([]byte(snsNotification.Message), &message)
	if err != nil {
		s.logger.Info().Str("message", snsNotification.Message).Msg("Error parsing message")
		s.logger.Error().Err(err)
		return err
	}
	err = s.eventService.CreateSESEvent(context.Background(), message)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error creating event")
		return err
	}

	return nil
}

func (s *snsService) parsePayload(payload []byte) (*constant.AwsRegion, *snsNotificationPayload, error) {
	s.logger.Info().Msg("Parsing SNS payload")
	var snsNotification snsNotificationPayload
	err := json.Unmarshal(payload, &snsNotification)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error parsing payload")
		return nil, nil, err
	}
	arnParts := strings.Split(snsNotification.TopicArn, ":")
	if len(arnParts) != 6 {
		return nil, nil, errors.New("invalid ARN")
	}
	region := constant.AwsRegion(arnParts[3])

	return &region, &snsNotification, nil
}

func (s *snsService) verifySignature(n snsNotificationPayload) error {
	s.logger.Info().Msg("Verifying signature")
	cert, err := s.getSigningCertificate(n.SigingCertURL)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error getting signing certificate")
		return err
	}
	signatureBytes, err := base64.StdEncoding.DecodeString(n.Signature)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error decoding signature")
		return err
	}
	s.logger.Info().Msg("Signature decoded")
	return cert.CheckSignature(x509.SHA1WithRSA, s.buildSignature(n), signatureBytes)
}

func (s *snsService) buildSignature(n snsNotificationPayload) []byte {
	var b bytes.Buffer
	signatureParts := []string{}
	signatureParts = append(signatureParts, "Message")
	signatureParts = append(signatureParts, n.Message)
	signatureParts = append(signatureParts, "MessageId")
	signatureParts = append(signatureParts, n.MessageId)
	// to handle notification where there is no mention of message
	if n.Subject != nil {
		signatureParts = append(signatureParts, "Subject")
		signatureParts = append(signatureParts, *n.Subject)
	}
	// to handle subscribe and unsubscribe, there is no mention of unsubscribe in the docs
	if n.SubscribeURL != nil {
		signatureParts = append(signatureParts, "SubscribeURL")
		signatureParts = append(signatureParts, *n.SubscribeURL)
	}
	signatureParts = append(signatureParts, "Timestamp")
	signatureParts = append(signatureParts, n.Timestamp)
	// to handle subscribe and unsubscribe
	if n.Token != nil {
		signatureParts = append(signatureParts, "Token")
		signatureParts = append(signatureParts, *n.Token)
	}
	signatureParts = append(signatureParts, "TopicArn")
	signatureParts = append(signatureParts, n.TopicArn)
	signatureParts = append(signatureParts, "Type")
	signatureParts = append(signatureParts, n.Type)

	for _, part := range signatureParts {
		b.WriteString(part + "\n")
	}

	return b.Bytes()
}

func (s *snsService) getSigningCertificate(certificateUrl string) (*x509.Certificate, error) {
	s.logger.Info().Str("certificateUrl", certificateUrl).Msg("Getting signing certificate")
	parsedURL, err := url.Parse(certificateUrl)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	existingCertificate, ok := s.signingCertificates[parsedURL.String()]
	s.mu.RUnlock()
	if ok {
		s.logger.Info().Msg("Certificate found in cache")
		return existingCertificate, nil
	}
	s.logger.Info().Msg("Certificate not found in cache, fetching from URL")
	resp, err := http.Get(certificateUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		s.logger.Error().Int("statusCode", resp.StatusCode).Msg("non 200 response on certificate URL")
		return nil, err
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(bytes)
	if block == nil {
		s.logger.Error().Msg("Invalid PEM")
		return nil, errors.New("invalid PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error parsing certificate")
		return nil, err
	}
	s.mu.Lock()
	s.signingCertificates[parsedURL.String()] = cert
	s.mu.Unlock()
	s.logger.Info().Msg("Certificate added to cache")

	return cert, nil
}
