package logic

import (
	"bluebell/dao/redis"
	"bluebell/models"
	"go.uber.org/zap"
	"strconv"
)

/*
投一票就加432分 86400/200 -> 200张赞成票就可以给帖子在首页续天  -> 《redis实战》

记录谁给哪个帖子投了什么票
更新文章分数：赞成票要加分；反对票减分

用户可能会进行投票：1.赞成票 2.反对票 3.不投 4.与上次投票相反
	v=1时，有两种情况
		1.之前没投过票，现在要投赞成票		--> 更新分数和投票记录		差值的绝对值：1  +432
		2.之前投过反对票，现在要改为赞成票	--> 更新分数和投票记录		差值的绝对值：2  +432*2
	v=0时，有两种情况
		1.之前投过反对票，现在要取消			--> 更新分数和投票记录		差值的绝对值：1  +432
		2.之前投过赞成票，现在要取消			--> 更新分数和投票记录		差值的绝对值：1  -432
	v=-1时，有两种情况
		1.之前没投过票，现在要投反对票		--> 更新分数和投票记录		差值的绝对值：1  -432
		2.之前投过赞成票，现在要改为反对票	--> 更新分数和投票记录		差值的绝对值：2  -432*2

投票限制：
每个帖子发表后一周内允许投票，超过后禁止
	1、到期之后将redis中保存的赞成票数及反对票数存储到mysql表中
	2、到期之后删除 KeyPostVotedZSetPrefix(存储帖子投票信息)
*/

// VoteForPost 投票功能
func VoteForPost(userId uint64, p *models.VoteDataForm) error {
	zap.L().Debug("VoteForPost",
		zap.Uint64("userId", userId),
		zap.String("postId", p.PostID),
		zap.Int8("Direction", p.Direction))
	return redis.VoteForPost(strconv.Itoa(int(userId)), p.PostID, float64(p.Direction))
}
