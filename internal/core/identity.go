package core

import (
	"context"
	"reflect"
	"strconv"

	"github.com/usesend0/send0/internal/uid"
)

type Identity interface {
	JTI() string
	UserId() uid.UID
	WorkspaceId() uid.UID
}

type identity struct {
	jti         string
	userId      uid.UID
	workspaceId uid.UID
}

type IdentityOptions struct {
	JTI         string
	Sub         string
	WorkspaceId *string
}

type identityContextKey struct{}

func NewIdentity(options IdentityOptions) (Identity, error) {
	identity := &identity{jti: options.JTI}
	userId, err := strconv.Atoi(options.Sub)
	if err != nil {
		return nil, err
	}
	identity.userId = *uid.NewUID(int64(userId))
	// Parse workspace ID if provided
	if options.WorkspaceId == nil {
		workspaceId, err := strconv.Atoi(*options.WorkspaceId)
		if err != nil {
			return nil, err
		}
		identity.workspaceId = *uid.NewUID(int64(workspaceId))
	}

	return identity, nil
}

func (u *identity) JTI() string {
	return u.jti
}

func (u *identity) UserId() uid.UID {
	return u.userId
}

func (u *identity) WorkspaceId() uid.UID {
	return u.workspaceId
}

func IdentityFromContext(ctx context.Context) Identity {
	ctxValue, ok := ctx.Value(identityContextKey{}).(Identity)
	if !ok {
		return nil
	}

	return ctxValue
}

func IdentityToContext(ctx context.Context, identity Identity) context.Context {
	if reflect.ValueOf(identity).IsNil() {
		return ctx
	}

	return context.WithValue(ctx, identityContextKey{}, identity)
}
