package service

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/usesend0/send0/internal/crypto"
	"github.com/usesend0/send0/internal/model"
	"github.com/usesend0/send0/internal/uid"
)

const webhookRequestSignatureBitSize = 1024
const maxWebhookLogRetries = 3

type WebhookService interface {
	Create(ctx context.Context, webhook *model.Webhook) (*model.Webhook, error)
	Get(ctx context.Context, id uid.UID) (*model.Webhook, error)
	List(ctx context.Context, workspaceId uid.UID) ([]*model.Webhook, error)
	StartListeners()
}

type webhookService struct {
	*baseService
	eventService EventSevice
	httpClient   *http.Client
}

func NewWebhookService(baseService *baseService, eventService EventSevice) WebhookService {
	return &webhookService{
		baseService,
		eventService,
		http.DefaultClient,
	}
}

func (s *webhookService) Create(ctx context.Context, webhook *model.Webhook) (*model.Webhook, error) {
	privateKey, publicKey, err := crypto.GenerateKeyPair(webhookRequestSignatureBitSize)
	if err != nil {
		return nil, err
	}
	webhook.SigningKey = model.ToJSONPrivateKey(*privateKey)
	webhook.SigningKeyPublic, err = crypto.PublicKeyToEncoded(publicKey)
	if err != nil {
		return nil, err
	}
	err = s.repository.Webhook.Save(ctx, webhook)

	return webhook, err
}

func (s *webhookService) Get(ctx context.Context, id uid.UID) (*model.Webhook, error) {
	return s.repository.Webhook.FindById(ctx, id)
}

func (s *webhookService) List(ctx context.Context, workspaceId uid.UID) ([]*model.Webhook, error) {
	return s.repository.Webhook.FindByWorkspaceId(ctx, workspaceId)
}

func (s *webhookService) CreateWebhookRequestSignature(
	signingKey *rsa.PrivateKey,
	id uid.UID,
	timestamp string,
	payload string,
) (*string, error) {
	sigatureContent := id.String() + timestamp + payload
	hashBytes, err := crypto.GenerateHash([]byte(sigatureContent))
	if err != nil {
		return nil, err
	}
	signature, err := crypto.SignHash(signingKey, hashBytes)
	if err != nil {
		return nil, err
	}
	// encode the signature to base64
	signatureEncoded := base64.StdEncoding.EncodeToString(signature)

	return &signatureEncoded, nil
}

func (s *webhookService) StartListeners() {
	cpus := runtime.NumCPU()
	for i := 0; i < cpus; i++ {
		go s.listen()
	}
}

func (s *webhookService) listen() {
	ch := s.eventService.Subscribe(fmt.Sprintf("WL-%s", s.uidGenerator.Next()))
	for event := range ch {
		s.handleEvent(event)
	}
}

func (s *webhookService) handleEvent(event *model.Event) {
	webhook, err := s.repository.Webhook.FindByEventType(context.Background(), event.EventType)
	if err != nil {
		s.logger.Error().Err(err).Msg("webhookService.handleEvent")
		return
	}
	bytes, err := json.Marshal(event)
	if err != nil {
		return
	}
	privateKey := model.FromJSONPrivateKey(webhook.SigningKey)
	webhookLog := model.WebhookEvent{
		Payload: string(bytes),
	}
	webhookRequest := model.WehbookRequest{
		SendAt: time.Now().UTC().Format(time.RFC3339),
	}
	signature, err := s.CreateWebhookRequestSignature(&privateKey, webhook.Id, webhookRequest.SendAt, webhookLog.Payload)
	if err != nil {
		return
	}
	webhookRequest.Signature = *signature
	webhookLog.Requests = model.JSONBArrayWebhookRequests{webhookRequest}
	err = execRetry(func() error {
		return s.repository.Webhook.SaveLog(context.Background(), &webhookLog)
	}, maxWebhookLogRetries)
	if err != nil {
		s.logger.Error().Err(err).Msg("max retries reached")
	}
	resp, status, err := s.postRequest(webhook.URL, bytes, map[string]string{
		"X-Webhook-Signature": webhookRequest.Signature,
		"X-Webhook-Id":        webhook.Id.String(),
		"X-Webhook-Timestamp": webhookRequest.SendAt,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("error sending webhook request")
		return
	}
	webhookLog.Requests[0].ResponseStatus = status
	webhookLog.Requests[0].ResponseBody = string(resp)

	// TODO: update webhook log with response status and body
	err = execRetry(func() error {
		return s.repository.Webhook.SaveLog(context.Background(), &webhookLog)
	}, maxWebhookLogRetries)
	if err != nil {
		s.logger.Error().Err(err).Msg("max retries reached")
	}
}

func (s *webhookService) postRequest(
	url string,
	payload []byte,
	headers map[string]string,
) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}
