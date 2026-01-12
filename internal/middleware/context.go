package middleware

import "context"

func GetUser(ctx context.Context) *UserClaims {
	user, ok := ctx.Value(UserContextKey).(*UserClaims)
	if !ok {
		return nil
	}
	return user
}
