package service

import (
	"context"
	"github.com/zmsocc/practice/webook/interactive/domain"
	"github.com/zmsocc/practice/webook/interactive/repository"
	"github.com/zmsocc/practice/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=interactive.go -package=svcmocks -destination=mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	CancelCollect(ctx context.Context, biz string, bizId, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.Logger
}

func NewInteractiveService(repo repository.InteractiveRepository, l logger.Logger) InteractiveService {
	return &interactiveService{
		repo: repo,
		l:    l,
	}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) Like(ctx context.Context, biz string, bizId, uid int64) error {
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	return i.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (i *interactiveService) CancelCollect(ctx context.Context, biz string, bizId, uid int64) error {
	return i.repo.DecrCollection(ctx, biz, bizId, uid)
}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	var (
		eg        errgroup.Group
		intr      domain.Interactive
		liked     bool
		collected bool
	)
	eg.Go(func() error {
		var err error
		intr, err = i.repo.Get(ctx, biz, bizId)
		return err
	})
	eg.Go(func() error {
		var err error
		liked, err = i.repo.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		var err error
		liked, err = i.repo.Collected(ctx, biz, bizId, uid)
		return err
	})
	err := eg.Wait()
	if err != nil {
		return domain.Interactive{}, err
	}
	intr.Liked = liked
	intr.Collected = collected
	return intr, err
}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	intrs, err := i.repo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}
