package service

import (
	"errors"

	"github.com/ozline/tiktok/cmd/chat/dal/db"
	"github.com/ozline/tiktok/cmd/chat/dal/mq"
	"github.com/ozline/tiktok/cmd/chat/rpc"
	"github.com/ozline/tiktok/kitex_gen/chat"
	"github.com/ozline/tiktok/kitex_gen/user"
)

func (c *ChatService) SendMessage(req *chat.MessagePostRequest, userId int64, createAt string) error {
	if len(req.Content) == 0 || len(req.Content) > 1000 {
		return errors.New("character limit error")
	}
	// ensure both sender and receiver exist to avoid FK errors
	if _, err := rpc.GetUser(c.ctx, &user.InfoRequest{UserId: userId, Token: req.Token}); err != nil {
		return err
	}
	if _, err := rpc.GetUser(c.ctx, &user.InfoRequest{UserId: req.ToUserId, Token: req.Token}); err != nil {
		return err
	}
	message := &mq.MiddleMessage{
		Id:         db.SF.NextVal(),
		ToUserId:   req.ToUserId,
		FromUserId: userId,
		Content:    req.Content,
		CreatedAt:  createAt,
	}
	if mq.ChatMQCli == nil {
		return mq.PersistMessage(c.ctx, message)
	}

	if err := mq.ChatMQCli.Publish(c.ctx, message); err != nil {
		// graceful degradation: process synchronously when Kafka is unavailable
		return mq.PersistMessage(c.ctx, message)
	}
	return nil
}
