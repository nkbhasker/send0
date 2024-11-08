package constant

import "github.com/aws/aws-sdk-go-v2/service/sesv2/types"

const (
	EventTypeEmailSend            EventType = "EMAIL_SEND"
	EventTypeEmailSendFailed      EventType = "EMAIL_SEND_FAILED"
	EventTypeEmailOpened          EventType = "EMAIL_OPENED"
	EventTypeEmailClicked         EventType = "EMAIL_CLICKED"
	EventTypeEmailBounced         EventType = "EMAIL_BOUNCED"
	EventTypeEmailReported        EventType = "EMAIL_REPORTED"
	EventTypeEmailRejected        EventType = "EMAIL_REJECTED"
	EventTypeEmailDelivered       EventType = "EMAIL_DELIVERED"
	EventTypeEmailUnsubsribed     EventType = "EMAIL_UNSUBSCRIBED"
	EventTypeEmailDeliveryDelayed EventType = "EMAIL_DELIVERY_DELAYED"
	EventTypeLinkClicked          EventType = "LINK_CLICKED"
	EventTypeOptIn                EventType = "OPT_IN"
)

type EventType string

var AwsSESEventTypeToEventType = map[types.EventType]EventType{
	types.EventTypeSend:          EventTypeEmailSend,
	types.EventTypeOpen:          EventTypeEmailOpened,
	types.EventTypeClick:         EventTypeEmailClicked,
	types.EventTypeBounce:        EventTypeEmailBounced,
	types.EventTypeComplaint:     EventTypeEmailReported,
	types.EventTypeReject:        EventTypeEmailRejected,
	types.EventTypeDelivery:      EventTypeEmailDelivered,
	types.EventTypeSubscription:  EventTypeEmailUnsubsribed,
	types.EventTypeDeliveryDelay: EventTypeEmailDeliveryDelayed,
}
