package service

import (
	"github.com/ozline/tiktok/cmd/user/dal/db"
	"github.com/ozline/tiktok/kitex_gen/user"
	"github.com/ozline/tiktok/pkg/errno"
	"golang.org/x/crypto/bcrypt"
)

// CheckUser check if user exists and it's password
func (s *UserService) CheckUser(req *user.LoginRequest) (*db.User, error) {
	// 函数名叫Check，却返回User? 应该返回 Bool
	userModel, err := db.GetUserByUsername(s.ctx, req.Username)

	if err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(req.Password)) != nil {
		return nil, errno.AuthorizationFailedError
	}

	return userModel, nil
}
