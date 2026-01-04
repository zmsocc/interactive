package domain

import "time"

type Interactive struct {
	Id         int64  `json:"id"`
	Biz        string `json:"biz"`
	BizId      int64  `json:"biz_id"`
	ReadCnt    int64  `json:"read_cnt"`
	LikeCnt    int64  `json:"like_cnt"`
	CollectCnt int64  `json:"collect_cnt"`
	Utime      time.Time
	Ctime      time.Time
	Liked      bool `json:"liked"`
	Collected  bool `json:"collected"`
}

type Self struct {
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

//type Collection struct {
//	Name string
//	Uid  int64
//}
