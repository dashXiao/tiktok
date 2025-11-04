package middleware

import (
	"context"

	"github.com/cloudwego/kitex/pkg/endpoint"
)

var _ endpoint.Middleware = CommonMiddleware

// server for api
func CommonMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		return next(ctx, req, resp)
	}
}
