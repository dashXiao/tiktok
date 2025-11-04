package cache

import (
	"context"

	"github.com/ozline/tiktok/cmd/interaction/dal/db"
	"github.com/ozline/tiktok/pkg/constants"
	"github.com/redis/go-redis/v9"
)

func GetComments(ctx context.Context, key string) (*[]redis.Z, error) {
	pipe := RedisClient.TxPipeline()
	commentKey := GetCommentKey(key)
	if err := pipe.TTL(ctx, commentKey).Err(); err != nil {
		return nil, err
	}
	if err := pipe.ZRevRangeWithScores(ctx, commentKey, 0, -1).Err(); err != nil {
		return nil, err
	}
	cmders, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	for _, cmder := range cmders {
		if err := cmder.Err(); err != nil {
			return nil, err
		}
	}
	lastTime := cmders[0].(*redis.DurationCmd).Val()
	rComments := cmders[1].(*redis.ZSliceCmd).Val()
	if lastTime < constants.CommentExpiredTime/2 {
		if err := RedisClient.Expire(ctx, commentKey, constants.CommentExpiredTime).Err(); err != nil {
			return nil, err
		}
	}
	return &rComments, nil
}

func AddComment(ctx context.Context, key string, comment *db.Comment) error {
	data, err := comment.MarshalMsg(nil)
	if err != nil {
		return err
	}
	return RedisClient.ZAdd(ctx, key, redis.Z{Score: float64(comment.CreatedAt.Unix()), Member: data}).Err()
}

func AddComments(ctx context.Context, key string, comments *[]db.Comment) error {
	commentKey := GetCommentKey(key)
	zComments := make([]redis.Z, len(*comments))
	for i := 0; i < len(*comments); i++ {
		data, err := (*comments)[i].MarshalMsg(nil)
		if err != nil {
			return err
		}
		zComments[i] = redis.Z{Score: float64((*comments)[i].CreatedAt.Unix()), Member: data}
	}
	pipe := RedisClient.TxPipeline()
	if err := pipe.ZAdd(ctx, commentKey, zComments...).Err(); err != nil {
		return err
	}
	if err := pipe.Expire(ctx, commentKey, constants.CommentExpiredTime).Err(); err != nil {
		return err
	}
	cmders, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	for _, cmder := range cmders {
		if err := cmder.Err(); err != nil {
			return err
		}
	}
	return nil
}

func AddNoData(ctx context.Context, key string) error {
	zData := redis.Z{}
	pipe := RedisClient.TxPipeline()
	commentKey := GetCommentKey(key)
	if err := pipe.ZAdd(ctx, commentKey, zData).Err(); err != nil {
		return err
	}
	if err := pipe.Expire(ctx, commentKey, constants.NoDataExpiredTime).Err(); err != nil {
		return err
	}
	cmders, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	for _, cmder := range cmders {
		if err := cmder.Err(); err != nil {
			return err
		}
	}
	return nil
}

func DeleteComment(ctx context.Context, key string, comment *db.Comment) error {
	data, err := comment.MarshalMsg(nil)
	if err != nil {
		return err
	}
	return RedisClient.ZRem(ctx, key, data).Err()
}

func GetCount(ctx context.Context, key string) (bool, string, error) {
	count, err := RedisClient.Get(ctx, GetCountKey(key)).Result()
	if err == redis.Nil {
		return false, count, nil
	}
	if err != nil {
		return true, count, err
	}
	return true, count, nil
}

func SetCount(ctx context.Context, key string, count int64) error {
	return RedisClient.Set(ctx, GetCountKey(key), count, constants.CommentExpiredTime).Err()
}

func IsExistComment(ctx context.Context, key string) (int64, error) {
	commentKey := GetCommentKey(key)
	return RedisClient.Exists(ctx, commentKey).Result()
}

func Delete(ctx context.Context, key string) error {
	return RedisClient.Del(ctx, key).Err()
}

func Unlink(ctx context.Context, key string) error {
	commentKey := GetCommentKey(key)
	return RedisClient.Unlink(ctx, commentKey).Err()
}

func Lock(ctx context.Context, key string) (bool, error) {
	return RedisClient.SetNX(ctx, key, 1, constants.LockTime).Result()
}

func AddCount(ctx context.Context, increment int64, videoID string) error {
	return RedisClient.IncrBy(ctx, GetCountKey(videoID), increment).Err()
}
