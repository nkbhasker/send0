package constant

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

const CustomDomainPrefix = "ses"

const (
	DomainStatusActive   DomainStatus = "ACTIVE"
	DomainStatusInactive DomainStatus = "INACTIVE"
	DomainStatusPending  DomainStatus = "PENDING"
)

const (
	DNSStatusActive     DNSStatus = "ACTIVE"
	DNSStatusPending    DNSStatus = "PENDING"
	DNSStatusIncomplete DNSStatus = "INCOMPLETE"
)

var _ sql.Scanner = (*JSONDomainRecords)(nil)
var _ driver.Valuer = (*JSONDomainRecords)(nil)

type DomainStatus string
type DNSStatus string
type domainRecord struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Type     string    `json:"type"`
	TTL      int       `json:"ttl"`
	Priority int       `json:"priority"`
	Status   DNSStatus `json:"status"`
}

type JSONDomainRecords []domainRecord

func (r *JSONDomainRecords) Scan(value interface{}) error {
	if value == nil {
		*r = []domainRecord{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, r)
	case string:
		return json.Unmarshal([]byte(v), r)
	default:
		return fmt.Errorf("invalid type: %T", value)
	}
}

func (r JSONDomainRecords) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]domainRecord{})
	}

	return json.Marshal(r)
}
