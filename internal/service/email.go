package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/model"
)

type EmailService interface {
	Send(ctx context.Context, requestId string, email []*model.Email) ([]string, error)
}

type emailService struct {
	*baseService
	sesService   SESService
	eventService EventSevice
}

func NewEmailService(baseService *baseService, sesService SESService, eventService EventSevice) EmailService {
	return &emailService{
		baseService:  baseService,
		sesService:   sesService,
		eventService: eventService,
	}
}

func (s *emailService) Send(ctx context.Context, requestId string, emails []*model.Email) ([]string, error) {
	emailIds := make([]string, 0, len(emails))
	err := s.Transact(ctx, func(ctx context.Context, service *Service) error {
		for _, email := range emails {
			email.Id = *s.uidGenerator.Next()
			email.RequestId = requestId
			email.EmailContent.EmailId = email.Id
			email.EmailContent.OrganizationId = email.OrganizationId
			email.EmailContent.WorkspaceId = email.WorkspaceId
			err := s.repository.Email.Save(ctx, email)
			if err != nil {
				return err
			}
			emailIds = append(emailIds, email.Id.String())
		}
		return nil
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed sending emails")
		return nil, errors.New("failed sending emails")
	}
	// TODO: send emails through SES, if it fails, update the status of the email to FAILED

	return emailIds, nil
}

func (s *emailService) SendEmails(ctx context.Context, emails []*model.Email) ([]string, error) {
	domains := make(map[string]*model.Domain)
	events := make([]*model.Event, 0)
	for _, email := range emails {
		err := func() error {
			fromAddress, err := mail.ParseAddress(email.From)
			if err != nil {
				return err
			}
			domainName := strings.Split(fromAddress.Address, "@")[1]
			domain, ok := domains[domainName]
			if !ok {
				domain, err = s.repository.Domain.FindByDomainName(ctx, email.WorkspaceId, email.OrganizationId, domainName)
				if err != nil {
					return err
				}
				if domain == nil || domain.Status != constant.DomainStatusActive {
					return errors.New("Domain not found or not active")
				}
				domains[domainName] = domain
			}
			messageId, err := s.sesService.SendEmail(
				ctx,
				domain.Region,
				domain.Id.String(),
				email.From,
				email.Recipients.Addresses(),
				email.CCRecipients.Addresses(),
				email.BCCRecipients.Addresses(),
				email.EmailContent.Subject,
				email.EmailContent.Html,
				email.EmailContent.Text,
			)
			if err != nil {
				return err
			}
			email.SentAt = time.Now().UTC().String()
			email.MessageId = *messageId
			err = s.repository.Email.Save(ctx, email)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			events = append(events, &model.Event{
				Receipients:    email.Recipients.Addresses(),
				CCRecipients:   email.CCRecipients.Addresses(),
				BCCRecipients:  email.BCCRecipients.Addresses(),
				EventType:      constant.EventTypeEmailSendFailed,
				OrganizationId: email.OrganizationId,
				WorkspaceId:    email.WorkspaceId,
				MetaData: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
	}
	// TODO: do something with the events
	fmt.Println(events)

	return nil, nil
}

func ParseRecipients(recipients []string) ([]model.Recipient, error) {
	var parsedRecipients []model.Recipient
	for _, recipient := range recipients {
		_, err := mail.ParseAddress(recipient)
		if err != nil {
			return nil, err
		}
		parsedRecipients = append(parsedRecipients, model.Recipient{
			Address: recipient,
		})
	}

	return parsedRecipients, nil
}
