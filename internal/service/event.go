package service

import (
	"context"
	"fmt"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/model"
)

const eventSaveMaxRetries = 3

type EventSevice interface {
	Create(ctx context.Context, event *model.Event)
	CreateSESEvent(ctx context.Context, message sesNotificationMessage) error
	StartListeners()
	Publish(event []*model.Event)
	Subscribe(id string) chan *model.Event
}

type eventService struct {
	*baseService
	subs map[string]chan *model.Event
}

func NewEventService(baseService *baseService) EventSevice {
	return &eventService{
		baseService,
		make(map[string]chan *model.Event),
	}
}

func (s *eventService) Create(ctx context.Context, event *model.Event) {
	for i := 0; i < eventSaveMaxRetries; i++ {
		err := s.repository.Event.Save(ctx, event)
		if err == nil {
			break
		}
	}

	s.logger.Error().Err(fmt.Errorf("failed to save event")).Msg("eventService.Create")
}

func (s *eventService) CreateSESEvent(ctx context.Context, message sesNotificationMessage) error {
	email, err := s.repository.Email.FindByMessageId(context.Background(), message.Mail.MessageId)
	if err != nil {
		return err
	}
	event := &model.Event{
		MetaData: model.EventMetaData{
			"Message": message,
		},
		OrganizationId: email.OrganizationId,
		WorkspaceId:    email.WorkspaceId,
	}
	switch message.EventType {
	case types.EventTypeBounce:
		event.EventType = constant.EventTypeEmailBounced
	}

	return s.repository.Event.Save(ctx, event)
}

func (s *eventService) StartListeners() {
	cpus := runtime.NumCPU()
	for i := 0; i < cpus; i++ {
		go s.listen()
	}
}

func (s *eventService) listen() {
	ch := s.Subscribe(fmt.Sprintf("EL-%s", s.uidGenerator.Next()))
	for event := range ch {
		fmt.Println(event)
	}
}

func (s *eventService) Publish(event []*model.Event) {
	for _, event := range event {
		for _, ch := range s.subs {
			ch <- event
		}
	}
}

func (s *eventService) Subscribe(id string) chan *model.Event {
	if ch, ok := s.subs[id]; ok {
		return ch
	}

	ch := make(chan *model.Event, 100)
	s.subs[id] = ch

	return ch
}
