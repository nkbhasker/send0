package model

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/usesend0/send0/internal/crypto"
	"github.com/usesend0/send0/internal/storage/cache"
	"github.com/usesend0/send0/internal/storage/db"
	"github.com/usesend0/send0/internal/uid"
)

const (
	TableNameAuthn        TableName = "authn"
	TableNameClient       TableName = "clients"
	TableNameDomain       TableName = "domains"
	TableNameEmail        TableName = "emails"
	TableNameEmailContent TableName = "email_contents"
	TableNameEvent        TableName = "events"
	TableNameOrganization TableName = "organizations"
	TableNameSNSTopic     TableName = "sns_topics"
	TableNameTag          TableName = "tags"
	TableNameTeam         TableName = "teams"
	TableNameTeamUser     TableName = "team_users"
	TableNameUser         TableName = "users"
	TableNameWorkspace    TableName = "workspaces"
)

const (
	DBTypeDomainStatus     = "domain_status"
	DBTypeCampaignStatus   = "campaign_status"
	DBTypeEventType        = "event_type"
	DBTypeIdentityProvider = "identity_provider"
	DBTypeTeamUserStatus   = "team_user_status"
)

var _ sql.Scanner = (*JSONBArray)(nil)
var _ driver.Valuer = (*JSONBArray)(nil)
var _ sql.Scanner = (*JSONBMap)(nil)
var _ driver.Valuer = (*JSONBMap)(nil)
var _ sql.Scanner = (*JSONPrivateKey)(nil)
var _ driver.Valuer = (*JSONPrivateKey)(nil)
var _ Transactioner[*Repository] = (*Repository)(nil)

type TableName string

type Transactioner[T interface{}] interface {
	Transact(context.Context, func(context.Context, T) error) error
}

type Repository struct {
	*baseRepository
	Authn        AuthnRepository
	Client       ClientRepository
	Domain       DomainRepository
	Email        EmailRepository
	Event        EventRepository
	Organization OrganizationRepository
	SNSTopic     SNSTopicRepository
	Team         TeamRepository
	Template     TemplateRepository
	User         UserRepository
	Webhook      WebhookRepository
	Workspace    WorkspaceRepository
}

type Base struct {
	Id        uid.UID `json:"id" gorm:"primaryKey;type:bigint;"`
	UpdatedAt *string `json:"updatedAt" db:"updated_at" gorm:"type:timestamp with time zone"`
}

type baseRepository struct {
	*zerolog.Logger
	cache.Cache
	db.DB
	uid.UIDGenerator
}

type JSONBArray []string
type JSONBMap map[string]interface{}
type JSONPrivateKey rsa.PrivateKey

func NewRepository(baseRepository *baseRepository) *Repository {
	return &Repository{
		baseRepository: baseRepository,
		Authn:          NewAuthnRepository(baseRepository),
		Client:         NewClientRepository(baseRepository),
		Domain:         NewDomainRepository(baseRepository),
		Email:          NewEmailRepository(baseRepository),
		Event:          NewEventRepository(baseRepository),
		Organization:   NewOrganizationRepository(baseRepository),
		SNSTopic:       NewSNSTopicRepository(baseRepository),
		Team:           NewTeamRepository(baseRepository),
		Template:       NewTemplateRepository(baseRepository),
		User:           NewUserRepository(baseRepository),
		Webhook:        NewWebhookRepository(baseRepository),
		Workspace:      NewWorkspaceRepository(baseRepository),
	}
}

func NewBaseRepository(
	cache cache.Cache,
	db db.DB,
	uidGenerator uid.UIDGenerator,
	logger *zerolog.Logger,
) *baseRepository {
	return &baseRepository{
		logger,
		cache,
		db,
		uidGenerator,
	}
}

func (r *baseRepository) Transact(ctx context.Context, fn func(ctx context.Context, repo *Repository) error) error {
	tx, db, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}
	baseRepository := NewBaseRepository(r.Cache, db, r.UIDGenerator, r.Logger)
	err = fn(ctx, NewRepository(baseRepository))
	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			r.Logger.Error().Err(rollbackErr).Msg("Failed to rollback transaction")
		}

		return err
	}

	return tx.Commit(ctx)
}

func (r *baseRepository) UID(id uid.UID) uid.UID {
	if id == (uid.UID{}) {
		return *r.UIDGenerator.Next()
	}

	return id
}

func (a *JSONBArray) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	default:
		return errors.New("invalid type")
	}
}

func (a JSONBArray) Value() (driver.Value, error) {
	if a == nil {
		return json.Marshal([]string{})
	}

	return json.Marshal(a)
}

func (m *JSONBMap) Scan(src interface{}) error {
	return scanJSONB(src, m)
}

func (m JSONBMap) Value() (driver.Value, error) {
	return valueJSONB(m)
}

func (p *JSONPrivateKey) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		privateKey, err := crypto.BytesToPrivateKey(v)
		if err != nil {
			return err
		}
		*p = JSONPrivateKey(*privateKey)
		return nil
	default:
		return errors.New("invalid type")
	}
}

func (p JSONPrivateKey) Value() (driver.Value, error) {
	// check if the private key is zero value
	if reflect.ValueOf(p).IsZero() {
		return nil, nil
	}
	privateKey := FromJSONPrivateKey(p)
	return crypto.PrivateKeyToBytes(&privateKey)
}

func ToJSONBArray(arr []string) JSONBArray {
	return JSONBArray(arr)
}

func FromJSONBArray(arr JSONBArray) []string {
	return []string(arr)
}

func ToJSONPrivateKey(key rsa.PrivateKey) JSONPrivateKey {
	return JSONPrivateKey(key)
}

func FromJSONPrivateKey(key JSONPrivateKey) rsa.PrivateKey {
	return rsa.PrivateKey(key)
}

func scanJSONB(src interface{}, dest interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, dest)
	default:
		return errors.New("invalid type")
	}
}

func valueJSONB(src interface{}) (driver.Value, error) {
	return json.Marshal(src)
}
