package session

import "context"

type User struct {
	ID    string `json:"id"`
	AppID int32  `json:"app_id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// userKeyType custom context key
type userKeyType struct{}

var userKey = userKeyType{}

func UserInCtx(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userKey).(*User)
	if !ok || user == nil {
		return nil, false
	}
	return user, true
}

func WithCtxUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserIDInCtx(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userKey).(*User)
	if !ok || user == nil {
		return "", false
	}
	return user.ID, true
}

func UserIDInCtxOrEmpty(ctx context.Context) string {
	id, _ := UserIDInCtx(ctx)
	return id
}

func AppIDInCtx(ctx context.Context) (int32, bool) {
	user, ok := ctx.Value(userKey).(*User)
	if !ok || user == nil {
		return 0, false
	}
	return user.AppID, true
}

func AppIDInCtxOrEmpty(ctx context.Context) int32 {
	appID, _ := AppIDInCtx(ctx)
	return appID
}
