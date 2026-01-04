package events

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/zmsocc/practice/webook/interactive/repository"
	"github.com/zmsocc/practice/webook/pkg/logger"
	"github.com/zmsocc/practice/webook/pkg/saramax"
	"time"
)

type InteractiveReadEventBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.Logger) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (k *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", k.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{"read_article"},
			saramax.NewBatchHandler[ReadEvent](k.l, k.Consume))
		if er != nil {
			k.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (k *InteractiveReadEventBatchConsumer) Consume(msg []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "article")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := k.repo.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		k.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Value: ids},
			logger.Error(err))
	}
	return nil
}

func t1() {
	var m []int
	m = make([]int, 3)
	n := []int{1, 2, 3}
	fmt.Println(m)
	fmt.Println(n)
}
