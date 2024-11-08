package service

import (
	"context"
	"errors"
	"strings"

	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/crypto"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/uid"
	"golang.org/x/net/publicsuffix"
)

const dkimSignatureBitSize = 1024

type DomainService interface {
	Create(ctx context.Context, domain *model.Domain) error
	Delete(ctx context.Context, domainId uid.UID) error
}

type domainService struct {
	*baseService
	ses SESService
}

func NewDomainService(baseService *baseService, sesService SESService) DomainService {
	return &domainService{
		baseService,
		sesService,
	}
}

func (s *domainService) Create(ctx context.Context, domain *model.Domain) error {
	name := domain.Name
	suffix, icann := publicsuffix.PublicSuffix(name)
	if !icann {
		return errors.New("invalid domain")
	}
	nameParts := strings.Split(name, ".")
	suffixParts := strings.Split(suffix, ".")
	// get the subdomain from the domain
	subdomain := strings.Join(nameParts[:len(nameParts)-len(suffixParts)-1], ".")
	privateKey, publicKey, err := crypto.GenerateKeyPair(dkimSignatureBitSize)
	if err != nil {
		return err
	}
	privateKeyBytes, err := crypto.PrivateKeyToBytes(privateKey)
	if err != nil {
		return err
	}
	publicKeyBytes, err := crypto.PublicKeyToBytes(publicKey)
	if err != nil {
		return err
	}
	domain = s.setupRecords(domain, subdomain, publicKeyBytes)
	domain.Id = *s.uidGenerator.Next()
	domain.Status = constant.DomainStatusPending
	domain.PrivateKey = model.ToJSONPrivateKey(*privateKey)
	err = s.repository.Domain.Save(ctx, domain)
	if err != nil {
		return err
	}
	err = s.ses.CreateEmailIdentity(ctx, domain, crypto.TrimPrefixAndSuffix(privateKeyBytes))
	if err != nil {
		return errors.New("failed to create domain")
	}

	return nil
}

func (s *domainService) Delete(ctx context.Context, domainId uid.UID) error {
	domain, err := s.repository.Domain.FindById(ctx, domainId)
	if err != nil {
		return err
	}
	if domain == nil {
		return errors.New("domain not found")
	}
	err = s.ses.DeleteEmailIdentity(ctx, domain)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to delete email identity")
		return errors.New("failed to delete domain")
	}

	return s.repository.Domain.Delete(ctx, domainId)
}

func (s *domainService) setupRecords(domain *model.Domain, subdomain string, publicKeyBytes []byte) *model.Domain {
	domain.DKIMRecords = constant.JSONDomainRecords{{
		Name:   formatSubdomain([]string{constant.AppName, "_domainkey", subdomain}),
		Value:  "p=" + crypto.TrimPrefixAndSuffix(publicKeyBytes),
		TTL:    300,
		Type:   "TXT",
		Status: constant.DNSStatusPending,
	}}
	domain.SPFRecords = constant.JSONDomainRecords{
		{
			Name:   formatSubdomain([]string{constant.CustomDomainPrefix, subdomain}),
			Value:  "v=spf1 include:amazonses.com -all",
			TTL:    300,
			Type:   "TXT",
			Status: constant.DNSStatusPending,
		},
		{
			Name:     formatSubdomain([]string{constant.CustomDomainPrefix, subdomain}),
			Value:    formatSubdomain([]string{"feedback-smtp", string(domain.Region), "amazonses.com"}),
			TTL:      300,
			Type:     "MX",
			Priority: 10,
			Status:   constant.DNSStatusPending,
		},
	}
	domain.DMARCRecords = constant.JSONDomainRecords{{
		Name:   formatSubdomain([]string{"_dmarc", subdomain}),
		Value:  "v=DMARC1; p=none;",
		TTL:    300,
		Type:   "TXT",
		Status: constant.DNSStatusPending,
	}}

	return domain
}
