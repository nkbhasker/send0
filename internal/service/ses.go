package service

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/model"
)

type sesService struct {
	*baseService
	svcs       map[constant.AwsRegion]*sesv2.Client
	snsService SNSService
}

type SESService interface {
	SendEmail(
		ctx context.Context,
		region constant.AwsRegion,
		configSetName string,
		from string,
		to []string,
		cc []string,
		bcc []string,
		subject *string,
		html *string,
		text *string,
	) (*string, error)
	CreateEmailIdentity(ctx context.Context, domain *model.Domain, privateKey string) error
	DeleteEmailIdentity(ctx context.Context, domain *model.Domain) error
}

func NewSESService(baseService *baseService, snsService SNSService) (SESService, error) {
	svcs := make(map[constant.AwsRegion]*sesv2.Client)
	for _, region := range constant.SupportedSESRegions {
		svcs[region] = sesv2.New(sesv2.Options{
			Region: string(region),
			Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
				baseService.config.SES.AccessKeyId,
				baseService.config.SES.SecretAccessKey,
				"",
			)),
		})
	}

	return &sesService{
		baseService: baseService,
		svcs:        svcs,
		snsService:  snsService,
	}, nil
}

func (s *sesService) SendEmail(
	ctx context.Context,
	region constant.AwsRegion,
	configSetName string,
	from string,
	to []string,
	cc []string,
	bcc []string,
	subject *string,
	html *string,
	text *string,
) (*string, error) {
	svc, ok := s.svcs[region]
	if !ok {
		return nil, fmt.Errorf("SES service not available for region %s", region)
	}
	if len(to) == 0 || len(cc) == 0 || len(bcc) == 0 {
		return nil, fmt.Errorf("no recipients provided")
	}
	message := &types.Message{
		Body: &types.Body{
			Html: &types.Content{
				Data: html,
			},
			Text: &types.Content{
				Data: text,
			},
		},
		Subject: &types.Content{
			Data: subject,
		},
	}
	resp, err := svc.SendEmail(ctx, &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses:  to,
			CcAddresses:  cc,
			BccAddresses: bcc,
		},
		Content: &types.EmailContent{
			Simple: message,
		},
		FromEmailAddress:     aws.String(from),
		ConfigurationSetName: aws.String(configSetName),
	})
	if err != nil {
		return nil, err
	}

	return resp.MessageId, nil
}

func (s *sesService) CreateEmailIdentity(ctx context.Context, domain *model.Domain, privateKey string) error {
	svc, ok := s.svcs[domain.Region]
	if !ok {
		return fmt.Errorf("SES service not available for region %s", domain.Region)
	}
	err := s.createConfigurationSet(domain)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create configuration set")
		return err
	}
	_, err = svc.CreateEmailIdentity(ctx, &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(domain.Name),
		DkimSigningAttributes: &types.DkimSigningAttributes{
			DomainSigningPrivateKey: aws.String(privateKey),
			DomainSigningSelector:   aws.String(constant.AppName),
		},
		ConfigurationSetName: aws.String(domain.Id.String()),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create email identity")
		return err
	}
	_, err = svc.PutEmailIdentityMailFromAttributes(ctx, &sesv2.PutEmailIdentityMailFromAttributesInput{
		EmailIdentity:  aws.String(domain.Name),
		MailFromDomain: aws.String(formatSubdomain([]string{constant.CustomDomainPrefix, domain.Name})),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create mail from domain")
		return err
	}
	return nil
}

func (s *sesService) createConfigurationSet(domain *model.Domain) error {
	svc, ok := s.svcs[domain.Region]
	if !ok {
		return fmt.Errorf("SES service not available for region %s", domain.Region)
	}
	_, err := svc.CreateConfigurationSet(context.Background(), &sesv2.CreateConfigurationSetInput{
		ConfigurationSetName: aws.String(domain.Id.String()),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create configuration set")
		return err
	}
	err = s.createConfigurationSetEventDestination(domain)
	if err != nil {
		return err
	}

	return nil
}

func (s *sesService) createConfigurationSetEventDestination(domain *model.Domain) error {
	svc, ok := s.svcs[domain.Region]
	if !ok {
		return fmt.Errorf("SES service not available for region %s", domain.Region)
	}
	topic, ok := s.snsService.Topics()[domain.Region]
	if !ok {
		return fmt.Errorf("SNS topic not available for region %s", domain.Region)
	}
	_, err := svc.CreateConfigurationSetEventDestination(context.Background(), &sesv2.CreateConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(domain.Id.String()),
		EventDestination: &types.EventDestinationDefinition{
			SnsDestination: &types.SnsDestination{
				TopicArn: aws.String(topic.Arn),
			},
			Enabled: true,
			MatchingEventTypes: []types.EventType{
				types.EventTypeSend,
				types.EventTypeBounce,
				types.EventTypeComplaint,
				types.EventTypeDelivery,
				types.EventTypeOpen,
				types.EventTypeClick,
				types.EventTypeReject,
				types.EventTypeDeliveryDelay,
			},
		},
		EventDestinationName: aws.String(constant.AppName),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create configuration set event destination")
		return err
	}

	return nil
}

func (s *sesService) DeleteEmailIdentity(ctx context.Context, domain *model.Domain) error {
	svc, ok := s.svcs[domain.Region]
	if !ok {
		return fmt.Errorf("SES service not available for region %s", domain.Region)
	}
	_, err := svc.DeleteEmailIdentity(ctx, &sesv2.DeleteEmailIdentityInput{
		EmailIdentity: aws.String(domain.Name),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to delete email identity")
		return err
	}
	// Delete the configuration set event destination
	err = s.deleteConfigurationSetEventDestination(domain)
	if err != nil {
		return err
	}
	// Delete the configuration set
	err = s.deleteConfigurationSet(domain)
	if err != nil {
		return err
	}
	return nil
}

func (s *sesService) deleteConfigurationSetEventDestination(domain *model.Domain) error {
	svc, ok := s.svcs[domain.Region]
	if !ok {
		return fmt.Errorf("SES service not available for region %s", domain.Region)
	}
	_, err := svc.DeleteConfigurationSetEventDestination(context.Background(), &sesv2.DeleteConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(domain.Id.String()),
		EventDestinationName: aws.String(constant.AppName),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to delete configuration set event destination")
		return err
	}

	return nil
}

func (s *sesService) deleteConfigurationSet(domain *model.Domain) error {
	svc, ok := s.svcs[domain.Region]
	if !ok {
		return fmt.Errorf("SES service not available for region %s", domain.Region)
	}
	_, err := svc.DeleteConfigurationSet(context.Background(), &sesv2.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(domain.Id.String()),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to delete configuration set")
		return err
	}
	return nil
}
