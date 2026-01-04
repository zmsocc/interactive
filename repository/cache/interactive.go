package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zmsocc/practice/webook/interactive/domain"
	"strconv"
	"time"
)

var (
	//go:embed lua/interactive_incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldCollectCnt = "collect_cnt"
	fieldLikeCnt    = "like_cnt"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
}

type RedisInteractiveCache struct {
	cmd redis.Cmdable
}

func NewRedisInteractiveCache(cmd redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{
		cmd: cmd,
	}
}

func (r *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldReadCnt, 1).Err()
}

func (r *RedisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldCollectCnt, 1).Err()
}

func (r *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldLikeCnt, 1).Err()
}

func (r *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldLikeCnt, -1).Err()
}

func (r *RedisInteractiveCache) DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldCollectCnt, -1).Err()
}

func (r *RedisInteractiveCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	key := r.key(biz, bizId)
	data, err := r.cmd.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(data) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	return domain.Interactive{
		Biz:        biz,
		BizId:      bizId,
		ReadCnt:    readCnt,
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
	}, err
}

func (r *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	key := r.key(biz, bizId)
	err := r.cmd.HSet(ctx, key, map[string]interface{}{
		fieldReadCnt:    intr.ReadCnt,
		fieldCollectCnt: intr.CollectCnt,
		fieldLikeCnt:    intr.LikeCnt,
	}).Err()
	if err != nil {
		return err
	}
	return r.cmd.Expire(ctx, key, time.Minute*15).Err()
}

func (r *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
