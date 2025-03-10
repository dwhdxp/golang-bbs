package redis

import (
	"bluebell_backend/models"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

// getIDsFormKey 按照score降序查询指定数量的帖子
func getIDsFormKey(key string, page, size int64) ([]string, error) {
	start := (page - 1) * size
	end := start + size - 1
	// ZRevRange 按照分数从大到小的顺序获取指定数量的元素
	return client.ZRevRange(key, start, end).Result()
}

// GetPostIDsInOrder 根据排序规则查询所有ids
func GetPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 1.根据用户请求中携带的order参数确定要查询的redisKey
	key := KeyPostTimeZSet            // 默认是时间
	if p.Order == models.OrderScore { // 按照分数请求
		key = KeyPostScoreZSet
	}
	// 2.查询ids范围 [(page-1)*size, (page-1)*size + size)
	return getIDsFormKey(key, p.Page, p.Size)
}

// GetPostVoteData 根据ids查询每篇帖子的赞成票数量
func GetPostVoteData(ids []string) (data []int64, err error) {
	data = make([]int64, 0, len(ids))
	for _, id := range ids {
		key := KeyPostVotedZSetPrefix + id
		// 统计每篇帖子的赞成票的数量
		// ZCount 返回有序集合中分数在[min, max]范围内的成员数量
		v := client.ZCount(key, "1", "1").Val()
		data = append(data, v)
	}
	// 使用 pipeline一次发送多条命令减少RTT
	//pipeline := client.Pipeline()
	//for _, id := range ids {
	//	key := KeyCommunityPostSetPrefix + id
	//	pipeline.ZCount(key, "1", "1")
	//}
	//cmders, err := pipeline.Exec()
	//if err != nil {
	//	return nil, err
	//}
	//data = make([]int64, 0, len(cmders))
	//for _, cmder := range cmders {
	//	v := cmder.(*redis.IntCmd).Val()
	//	data = append(data, v)
	//}
	return data, nil
}

// GetPostVoteNum 根据id查询每篇帖子的投赞成票的数据
func GetPostVoteNum(ids int64) (data int64, err error) {
	key := KeyPostVotedZSetPrefix + strconv.Itoa(int(ids))
	// 查找key中分数是1的元素数量 -> 统计每篇帖子的赞成票的数量
	data = client.ZCount(key, "1", "1").Val()
	return data, nil
}

// GetCommunityPostIDsInOrder  根据order查询community_id社区的ids
func GetCommunityPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 1.根据用户请求中携带的order参数确定要查询的redis key
	orderkey := KeyPostTimeZSet       // 默认是时间
	if p.Order == models.OrderScore { // 按照分数请求
		orderkey = KeyPostScoreZSet
	}

	// 使用ZInterStore 将存储某社区下所有帖子ID的Set 与 存储所有帖子得分信息的ZSet 合并生成一个新的ZSet
	// 新的ZSet存储的就是该社区下所有帖子得分信息

	// 社区的key
	cKey := KeyCommunityPostSetPrefix + strconv.Itoa(int(p.CommunityID))

	// 利用缓存key减少ZInterStore执行的次数 缓存key
	key := orderkey + strconv.Itoa(int(p.CommunityID))
	if client.Exists(key).Val() < 1 {
		// 不存在，需要计算
		pipeline := client.Pipeline()
		pipeline.ZInterStore(key, redis.ZStore{
			Aggregate: "MAX", // 将两个ZSet函数聚合的时候 求最大值
		}, cKey, orderkey) // ZInterStore 计算
		pipeline.Expire(key, 60*time.Second) // 设置超时时间
		_, err := pipeline.Exec()
		if err != nil {
			return nil, err
		}
	}
	// 存在的就直接根据key查询ids
	return getIDsFormKey(key, p.Page, p.Size)
}
