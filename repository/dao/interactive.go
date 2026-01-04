package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	DecrLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	DecrCollectInfo(ctx context.Context, cb UserCollectionBiz) error
	GetCollectInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error)
	GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error)
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error
	GetByIds(ctx context.Context, biz string, bizIds []int64) ([]Interactive, error)
}

type interactiveDAO struct {
	db *gorm.DB
}

func NewInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &interactiveDAO{
		db: db,
	}
}

func (d *interactiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		//Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"read_cnt": gorm.Expr("read_cnt + ?", 1),
			"utime":    now,
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Utime:   now,
		Ctime:   now,
	}).Error
}

func (d *interactiveDAO) IncrLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Biz:    biz,
			BizId:  bizId,
			Uid:    uid,
			Utime:  now,
			Ctime:  now,
			Status: 1,
		}).Error
		if err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("like_cnt + ?", 1),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   bizId,
			LikeCnt: 1,
			Utime:   now,
			Ctime:   now,
		}).Error
	})
}

func (d *interactiveDAO) DecrLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).
			Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).Where("biz = ? AND biz_id = ?", biz, bizId).
			Updates(map[string]interface{}{
				"utime":    now,
				"like_cnt": gorm.Expr("like_cnt + ?", -1),
			}).Error
	})
}

func (d *interactiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Utime = now
	cb.Ctime = now
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 插入收藏项目
		err := d.db.WithContext(ctx).Create(&cb).Error
		if err != nil {
			return err
		}
		// 这边就是更新数量
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("collect_cnt + ?", 1),
				"utime":       now,
			}),
		}).Create(&Interactive{
			Biz:        cb.Biz,
			BizId:      cb.BizId,
			CollectCnt: 1,
			Utime:      now,
			Ctime:      now,
		}).Error
	})
}

func (d *interactiveDAO) DecrCollectInfo(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Find(&UserLikeBiz{}).
			Where("biz = ? AND biz_id = ? AND uid = ?", cb.Biz, cb.BizId, cb.Uid).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).Where("biz = ? AND biz_id = ?", cb.Biz, cb.BizId).
			Updates(map[string]interface{}{
				"collect_cnt": gorm.Expr("collect_cnt + ?", -1),
				"utime":       now,
			}).Error
	})

}

func (d *interactiveDAO) GetCollectInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := d.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ? AND status = ?",
			biz, bizId, uid, 1).First(&res).Error
	return res, err
}

func (d *interactiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := d.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).First(&res).Error
	return res, err
}

func (d *interactiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := d.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ?", biz, bizId).First(&res).Error
	return res, err
}

func (d *interactiveDAO) BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error {
	// 可以用 map 合并吗？
	// 看情况。如果一批次里面，biz 和 bizId 都相等的占很多，那么就 map 合并，性能会更好
	// 不然你合并了没效果

	// 这个批次处理为什么比一次次提交要快？
	// A：十条消息调用十次 IncrReadCnt
	// B 就是批量
	// 事务本身的开销，A 是 B 的十倍
	// 刷新 redolog，undolog，binlog 到磁盘，A 是十次，B 是一次
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewInteractiveDAO(tx)
		for i := range biz {
			err := txDAO.IncrReadCnt(ctx, biz[i], bizId[i])
			if err != nil {
				// 记个日志
				// 也可以 return err
				return err
			}
		}
		return nil
	})
}

func (d *interactiveDAO) GetByIds(ctx context.Context, biz string, bizIds []int64) ([]Interactive, error) {
	var res []Interactive
	err := d.db.WithContext(ctx).Where("biz = ? AND id IN ?", biz, bizIds).Find(&res).Error
	return res, err
}

type UserLikeBiz struct {
	Id    int64  `gorm:"primaryKey;autoIncrement"`
	Biz   string `gorm:"uniqueIndex:uid_biz_id_type; type:varchar(128)"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	// 用户信息，谁点的赞
	Uid   int64 `gorm:"uniqueIndex:uid_biz_id_type"`
	Utime int64
	Ctime int64
	// 软删除
	Status uint8
}

type UserCollectionBiz struct {
	Id    int64  `gorm:"primaryKey;autoIncrement"`
	Biz   string `gorm:"uniqueIndex:uid_biz_id_type; type:varchar(128)"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Uid   int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Cid   int64
	Utime int64
	Ctime int64
	// 软删除
	Status uint8
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Utime      int64
	Ctime      int64
}

type Collection struct {
	Id    int64  `gorm:"primaryKey;autoIncrement"`
	Name  string `gorm:"type:varchar(1024)"`
	Uid   int64  `gorm:""`
	Ctime int64
	Utime int64
}
