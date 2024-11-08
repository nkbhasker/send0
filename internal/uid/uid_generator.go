package uid

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/sony/sonyflake"
	"github.com/usesend0/send0/internal/config"
	"github.com/usesend0/send0/internal/constant"
)

var startTime = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

type UIDGenerator interface {
	Next() *UID
}

type uidGenerator struct {
	*zerolog.Logger
	*sonyflake.Sonyflake
}

func NewUIDGenerator(cfg *config.Config, logger *zerolog.Logger) UIDGenerator {
	settings := sonyflake.Settings{
		StartTime: startTime,
	}
	if cfg.Env == constant.EnvDevelopment {
		settings.MachineID = func() (uint16, error) {
			return 1, nil
		}
	}

	return &uidGenerator{
		logger,
		sonyflake.NewSonyflake(settings),
	}
}

func (i *uidGenerator) Next() *UID {
	uid, err := i.NextID()
	if err != nil {
		i.Error().Err(err).Msg("Failed to generate UID")
		return i.Next()
	}

	return NewUID(int64(uid))
}

func Timestamp(uid int64) time.Time {
	return timestamp(uint64(uid))
}

func timestamp(uid uint64) time.Time {
	elapsedTime := sonyflake.ElapsedTime(uid)
	timestamp := startTime.Add(elapsedTime)

	return timestamp
}
