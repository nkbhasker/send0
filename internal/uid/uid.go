package uid

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var _ sql.Scanner = (*UID)(nil)
var _ driver.Valuer = (*UID)(nil)

type UID struct {
	id        int64
	timestamp time.Time
}

func (u *UID) Scan(src interface{}) error {
	switch v := src.(type) {
	case int64:
		*u = *NewUID(v)
	case []byte:
		id, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*u = *NewUID(id)
	default:
		return errors.New("incompatible type for uid")
	}

	return nil
}

func (u UID) Value() (driver.Value, error) {
	return u.id, nil
}

func NewUID(id int64) *UID {
	return &UID{
		id:        id,
		timestamp: timestamp(uint64(id)),
	}
}

func NewUIDFromString(id string) (*UID, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	return NewUID(int64(i)), nil
}

func (u UID) ID() int64 {
	return u.id
}

func (u UID) Timestamp() time.Time {
	return u.timestamp
}

func (u UID) String() string {
	return strconv.FormatInt(u.id, 10)
}

func (u UID) MarshalJSON() ([]byte, error) {
	if u.id == 0 {
		return json.Marshal(nil)
	}
	return json.Marshal(u.String())
}

func (u *UID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	*u = *NewUID(id)

	return nil
}

func (UID) GormDataType() string {
	return "bigint"
}

func (UID) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "bigint"
}
