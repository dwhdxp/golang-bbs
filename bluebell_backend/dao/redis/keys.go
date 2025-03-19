package redis

// redis key 注意使用命名空间的方式，方便查询和拆分
const (
	KeyPostInfoHashPrefix = "bluebell:post:"      // 存储帖子详细信息 Hash
	KeyPostTimeZSet       = "bluebell:post:time"  // 存储帖子发布时间信息 ZSet
	KeyPostScoreZSet      = "bluebell:post:score" // 存储帖子得分信息 ZSet
	//KeyPostVotedUpSetPrefix   = "bluebell:post:voted:down:"
	//KeyPostVotedDownSetPrefix = "bluebell:post:voted:up:"
	KeyPostVotedZSetPrefix    = "bluebell:post:voted:" // 存储某帖子投票信息 ZSet;后跟参数是post_id
	KeyCommunityPostSetPrefix = "bluebell:community:"  // 存储某社区下所有帖子ID Set;后跟参数community_id
)
