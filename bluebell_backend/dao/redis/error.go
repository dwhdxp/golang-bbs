package redis

import "errors"

var (
	ErrorVoteTimeExpire = errors.New("超过投票时间")
	ErrorVoted          = errors.New("已投票")
	ErrVoteRepeated     = errors.New("不允许重复投票")
)
