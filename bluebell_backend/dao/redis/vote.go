package redis

import (
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	OneWeekInSeconds          = 7 * 24 * 3600        // 一周的秒数
	OneMonthInSeconds         = 4 * OneWeekInSeconds // 一个月的秒数
	VoteScore         float64 = 432                  // 每一票的值432分
	PostPerAge                = 20                   // 每页显示20条帖子
)

// VoteForPost	为帖子投票
func VoteForPost(userID string, postID string, v float64) (err error) {
	// 1.投票限制：判断帖子发布是否超过一周
	postTime := client.ZScore(KeyPostTimeZSet, postID).Val()    // 从redis获取帖子发布时间
	if float64(time.Now().Unix())-postTime > OneWeekInSeconds { // 超过一个星期就不允许投票了
		// 不允许投票了
		return ErrorVoteTimeExpire
	}
	// 2.更新帖子的分数
	// 查询用户给帖子的投票记录，判断是否已投过票
	key := KeyPostVotedZSetPrefix + postID
	ov := client.ZScore(key, userID).Val() // 查询之前的投票记录
	if v == ov {
		return ErrVoteRepeated
	}
	// op为符号，判断增加/减少分数
	var op float64
	if v > ov {
		op = 1
	} else {
		op = -1
	}
	diffAbs := math.Abs(ov - v)                // 计算两次投票的差值
	pipeline := client.TxPipeline()            // 事务操作
	incrementScore := VoteScore * diffAbs * op // 计算更新分数
	// ZIncrBy 用于将有序集合中的成员分数增加指定数量
	_, err = pipeline.ZIncrBy(KeyPostScoreZSet, incrementScore, postID).Result()
	if err != nil {
		return err
	}
	// 3.记录用户为该帖子投票的数据
	if v == 0 {
		_, err = client.ZRem(key, postID).Result()
	} else {
		pipeline.ZAdd(key, redis.Z{ // 记录已投票
			Score:  v, // 赞成票还是反对票
			Member: userID,
		})
	}
	// 4.更新帖子的投票数
	pipeline.HIncrBy(KeyPostInfoHashPrefix+postID, "votes", int64(op))

	//switch math.Abs(ov) - math.Abs(v) {
	//case 1:
	//	// 取消投票 ov=1/-1 v=0
	//	// 投票数-1
	//	pipeline.HIncrBy(KeyPostInfoHashPrefix+postID, "votes", -1)
	//case 0:
	//	// 反转投票 ov=-1/1 v=1/-1
	//	// 投票数不用更新
	//case -1:
	//	// 新增投票 ov=0 v=1/-1
	//	// 投票数+1
	//	pipeline.HIncrBy(KeyPostInfoHashPrefix+postID, "votes", 1)
	//default:
	//	// 已经投过票了
	//	return ErrorVoted
	//}
	_, err = pipeline.Exec()
	return err
}

// CreatePost redis存储帖子相关信息
func CreatePost(postID, userID uint64, title, summary string, CommunityID uint64) (err error) {
	now := float64(time.Now().Unix())
	votedKey := KeyPostVotedZSetPrefix + strconv.Itoa(int(postID))             // bluebell:post:voted:post_id
	communityKey := KeyCommunityPostSetPrefix + strconv.Itoa(int(CommunityID)) // bluebell:community:community_id
	postInfo := map[string]interface{}{
		"title":    title,
		"summary":  summary,
		"post:id":  postID,
		"user:id":  userID,
		"time":     now,
		"votes":    1,
		"comments": 0,
	}

	// 事务操作：确保所有 Redis 操作要么全部成功，要么全部失败
	pipeline := client.TxPipeline()
	// 存储帖子投票信息 ZSet [bluebell:post:voted:post_id, (userID, score)]
	pipeline.ZAdd(votedKey, redis.Z{
		Score:  1, // 作者默认投赞成票
		Member: userID,
	})
	pipeline.Expire(votedKey, time.Second*OneMonthInSeconds*6) // 过期时间为6个月
	// 存储帖子得分信息 ZSet [bluebell:post:score, (post_id, score)]
	pipeline.ZAdd(KeyPostScoreZSet, redis.Z{
		Score:  now + VoteScore,
		Member: postID,
	})
	// 存储帖子发布时间信息 ZSet [bluebell:post:time, (post_id, score)]
	pipeline.ZAdd(KeyPostTimeZSet, redis.Z{
		Score:  now,
		Member: postID,
	})
	// 存储帖子详细信息 Hash [bluebell:post:post_id, postInfo]
	pipeline.HMSet(KeyPostInfoHashPrefix+strconv.Itoa(int(postID)), postInfo)
	// 存储某社区下所有帖子ID Set [bluebell:community:community_id, post_id]
	pipeline.SAdd(communityKey, postID)
	_, err = pipeline.Exec()
	return
}

// GetPost 从key中分页取出帖子
func GetPost(order string, page int64) []map[string]string {
	key := KeyPostScoreZSet
	if order == "time" {
		key = KeyPostTimeZSet
	}
	start := (page - 1) * PostPerAge
	end := start + PostPerAge - 1
	ids := client.ZRevRange(key, start, end).Val()
	postList := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		postData := client.HGetAll(KeyPostInfoHashPrefix + id).Val()
		postData["id"] = id
		postList = append(postList, postData)
	}
	return postList
}

// GetCommunityPost 分社区根据发帖时间或者分数取出分页的帖子
func GetCommunityPost(communityName, orderKey string, page int64) []map[string]string {
	key := orderKey + communityName // 创建缓存键

	if client.Exists(key).Val() < 1 {
		client.ZInterStore(key, redis.ZStore{
			Aggregate: "MAX",
		}, KeyCommunityPostSetPrefix+communityName, orderKey)
		client.Expire(key, 60*time.Second)
	}
	return GetPost(key, page)
}

// Reddit Hot rank algorithms
// from https://github.com/reddit-archive/reddit/blob/master/r2/r2/lib/db/_sorts.pyx
func Hot(ups, downs int, date time.Time) float64 {
	s := float64(ups - downs)
	order := math.Log10(math.Max(math.Abs(s), 1))
	var sign float64
	if s > 0 {
		sign = 1
	} else if s == 0 {
		sign = 0
	} else {
		sign = -1
	}
	seconds := float64(date.Second() - 1577808000)
	return math.Round(sign*order + seconds/43200)
}
