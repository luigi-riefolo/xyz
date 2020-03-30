package api

import (
	"context"

	"firebase.google.com/go/auth"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	emailKey             = "email"
	firebaseUserIDHeader = "firebase-user-id"
	userEmailHeader      = "user-email"
)

// BearerAuthUnaryClientInterceptor checks that each API
// request has a valid API token and is authenticated.
func BearerAuthUnaryClientInterceptor(authClient *auth.Client) grpc.UnaryClientInterceptor {

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		md := metautils.ExtractOutgoing(ctx)
		ctx = md.ToIncoming(ctx)

		// example: Authorization: Bearer YWxhZGRpbjpvcGVuc2VzYW1l
		token, err := grpc_auth.AuthFromMD(ctx, bearerKey)
		if err != nil {
			return errors.Wrap(err, "could not process authorization header")
		}

		// check that API token is valid
		t, err := authClient.VerifyIDTokenAndCheckRevoked(ctx, token)
		if err != nil {
			return errors.Wrap(err, "invalid API token")
		}

		md.Add(firebaseUserIDHeader, t.UID)
		md.Add(userEmailHeader, t.Claims[emailKey].(string))

		return invoker(md.ToOutgoing(ctx), method, req, reply, cc, opts...)
	}
}
