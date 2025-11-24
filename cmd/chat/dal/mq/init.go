package mq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/ozline/tiktok/cmd/chat/dal/cache"
	"github.com/ozline/tiktok/cmd/chat/dal/db"
	"github.com/ozline/tiktok/config"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm/clause"
)

type MiddleMessage struct {
	Id         int64
	ToUserId   int64
	FromUserId int64
	Content    string
	CreatedAt  string
}

type ChatMQ struct {
	writer    *kafka.Writer
	reader    *kafka.Reader
	dlqWriter *kafka.Writer
	topic     string
}

var (
	ChatMQCli *ChatMQ
	initOnce  sync.Once
	Mu        sync.Mutex
)

func Init() {
	initOnce.Do(func() {
		cfg := config.Kafka
		if cfg == nil {
			log.Println("[mq] kafka config not found")
			return
		}
		if len(cfg.Brokers) == 0 || cfg.Topic == "" {
			log.Println("[mq] kafka config missing brokers or topic")
			return
		}
		groupID := cfg.GroupID
		if groupID == "" {
			groupID = "chat-service"
		}

		dialer := &kafka.Dialer{
			Timeout:   5 * time.Second,
			DualStack: true,
			ClientID:  fmt.Sprintf("%s-chat", config.Server.Name),
		}

		writer := &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Brokers...),
			Topic:                  cfg.Topic,
			AllowAutoTopicCreation: true,
			Balancer:               &kafka.Hash{},
		}

		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			GroupID:        groupID,
			Topic:          cfg.Topic,
			MinBytes:       10e3,
			MaxBytes:       10e6,
			CommitInterval: time.Second,
			StartOffset:    kafka.LastOffset,
			Dialer:         dialer,
		})

		var dlqWriter *kafka.Writer
		if cfg.DeadLetterTopic != "" {
			dlqWriter = &kafka.Writer{
				Addr:                   kafka.TCP(cfg.Brokers...),
				Topic:                  cfg.DeadLetterTopic,
				AllowAutoTopicCreation: true,
				Balancer:               &kafka.Hash{},
			}
		}

		ChatMQCli = &ChatMQ{
			writer:    writer,
			reader:    reader,
			dlqWriter: dlqWriter,
			topic:     cfg.Topic,
		}

		go ChatMQCli.consume()
	})
}

func (c *ChatMQ) Publish(ctx context.Context, message *MiddleMessage) error {
	if c == nil || c.writer == nil {
		return errors.New("kafka writer is not initialized")
	}
	payload, err := sonic.Marshal(message)
	if err != nil {
		return err
	}

	return c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(compoundKey(message.FromUserId, message.ToUserId)),
		Value: payload,
		Time:  time.Now(),
	})
}

func (c *ChatMQ) consume() {
	if c == nil || c.reader == nil {
		return
	}
	ctx := context.Background()
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("[mq] fetch message failed: %v", err)
			continue
		}

		if err := c.handleMessage(ctx, msg.Value); err != nil {
			log.Printf("[mq] handle message failed: %v", err)
			c.pushToDLQ(ctx, msg, err)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[mq] commit offset failed: %v", err)
		}
	}
}

func (c *ChatMQ) handleMessage(ctx context.Context, payload []byte) error {
	middle := new(MiddleMessage)
	if err := sonic.Unmarshal(payload, middle); err != nil {
		return fmt.Errorf("unmarshal kafka payload: %w", err)
	}
	return PersistMessage(ctx, middle)
}

func (c *ChatMQ) pushToDLQ(ctx context.Context, msg kafka.Message, handleErr error) {
	if c.dlqWriter == nil {
		return
	}
	body, err := sonic.Marshal(map[string]interface{}{
		"error":   handleErr.Error(),
		"payload": string(msg.Value),
		"ts":      time.Now().UnixMilli(),
	})
	if err != nil {
		log.Printf("[mq] marshal dlq payload failed: %v", err)
		return
	}
	if err := c.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: body,
	}); err != nil {
		log.Printf("[mq] publish dlq message failed: %v", err)
	}
}

func PersistMessage(ctx context.Context, middle *MiddleMessage) error {
	if ctx == nil {
		ctx = context.Background()
	}
	cacheMessage := new(cache.Message)
	if err := convertForMysql(cacheMessage, middle); err != nil {
		return err
	}

	Mu.Lock()
	defer Mu.Unlock()

	if err := db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(cacheMessage).Error; err != nil {
		return err
	}

	key := compoundKey(cacheMessage.FromUserId, cacheMessage.ToUserId)
	revKey := compoundKey(cacheMessage.ToUserId, cacheMessage.FromUserId)
	payload, err := sonic.Marshal(middle)
	if err != nil {
		return err
	}

	return cache.MessageInsert(ctx, key, revKey, cacheMessage.CreatedAt.UnixMilli(), string(payload))
}

func compoundKey(from, to int64) string {
	return fmt.Sprintf("%d-%d", from, to)
}
