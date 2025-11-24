package service

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/ozline/tiktok/cmd/chat/dal/cache"
	"github.com/ozline/tiktok/cmd/chat/dal/db"
	"github.com/ozline/tiktok/cmd/chat/dal/mq"
	"github.com/ozline/tiktok/kitex_gen/chat"
)

// Get Messages history list
func (c *ChatService) GetMessages(req *chat.MessageListRequest, user_id int64) ([]*db.Message, error) {
	mq.Mu.Lock()
	defer mq.Mu.Unlock()
	// MySQL search
	messages, err := db.GetMessageList(c.ctx, req.ToUserId, user_id)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return messages, nil
	}

	sort.Sort(db.MessageArray(messages))

	for _, val := range messages {
		payload := &mq.MiddleMessage{
			Id:         val.Id,
			ToUserId:   val.ToUserId,
			FromUserId: val.FromUserId,
			Content:    val.Content,
			CreatedAt:  val.CreatedAt.Format(time.RFC3339),
		}
		mes, err := sonic.Marshal(payload)
		if err != nil {
			continue
		}
		cacheKey := strconv.FormatInt(val.FromUserId, 10) + "-" + strconv.FormatInt(val.ToUserId, 10)
		reverseKey := strconv.FormatInt(val.ToUserId, 10) + "-" + strconv.FormatInt(val.FromUserId, 10)
		creTime := val.CreatedAt.UnixMilli()
		if err := cache.MessageInsert(context.Background(), cacheKey, reverseKey, creTime, string(mes)); err != nil {
			continue
		}
	}
	return messages, nil
}
