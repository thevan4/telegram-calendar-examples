package main

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// grpc
var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

func newGRPCAuthInterceptor() grpc.UnaryServerInterceptor {
	isValid := func(authorization []string) bool {
		if len(authorization) < 1 {
			return false
		}
		token := strings.TrimPrefix(authorization[0], "Bearer ")
		return token == secretBearerToken
	}

	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if secretBearerToken != "" {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, errMissingMetadata
			}
			if !isValid(md["authorization"]) {
				return nil, errInvalidToken
			}
		}

		return handler(ctx, req)
	}
}

// http

func authHTTPInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if secretBearerToken != "" {
			authHeader := req.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != secretBearerToken {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, req)
	})
}
