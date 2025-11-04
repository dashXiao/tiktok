package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/ozline/tiktok/cmd/api/biz/pack"
	"github.com/ozline/tiktok/pkg/constants"
	"github.com/ozline/tiktok/pkg/errno"
	"github.com/ozline/tiktok/pkg/utils"
)

func AuthToken() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token, ok := c.GetQuery("token")

		if !ok {
			token, ok = c.GetPostForm("token")
		}

		if !ok {
			pack.SendFailResponse(c, errno.AuthorizationFailedError)
			c.Abort()
			return
		}

		claims, err := utils.CheckToken(token)

		if err != nil {
			pack.SendFailResponse(c, errno.AuthorizationFailedError)
			c.Abort()
			return
		}

		if claims.UserId < constants.StartID {
			pack.SendFailResponse(c, errno.AuthorizationFailedError)
			c.Abort()
		}
		// c.Set("current_user_id", claims.UserId)

		c.Next(ctx)
	}
}
