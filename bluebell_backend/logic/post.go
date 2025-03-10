package logic

import (
	"bluebell_backend/dao/mysql"
	"bluebell_backend/dao/redis"
	"bluebell_backend/models"
	"bluebell_backend/pkg/snowflake"
	"fmt"
	"strconv"

	"go.uber.org/zap"
)

// CreatePost 创建帖子
func CreatePost(post *models.Post) (err error) {
	// 1.根据雪花算法生成post_id(帖子ID)
	postID, err := snowflake.GetID()
	if err != nil {
		zap.L().Error("snowflake.GetID() failed", zap.Error(err))
		return
	}
	post.PostID = postID
	// 2.插入数据库
	if err := mysql.CreatePost(post); err != nil {
		zap.L().Error("mysql.CreatePost(&post) failed", zap.Error(err))
		return err
	}
	community, err := mysql.GetCommunityNameByID(fmt.Sprint(post.CommunityID))
	if err != nil {
		zap.L().Error("mysql.GetCommunityNameByID failed", zap.Error(err))
		return err
	}

	// 3.redis缓存帖子信息
	if err := redis.CreatePost(
		post.PostID,
		post.AuthorId,
		post.Title,
		TruncateByWords(post.Content, 120),
		community.CommunityID); err != nil {
		zap.L().Error("redis.CreatePost failed", zap.Error(err))
		return err
	}
	return
}

// GetPostById 根据Id查询帖子详情
func GetPostById(postID int64) (data *models.ApiPostDetail, err error) {
	// 1.查询帖子信息，根据post_id
	post, err := mysql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostByID(postID) failed",
			zap.Int64("postID", postID),
			zap.Error(err))
		return nil, err
	}
	// 2.查询作者信息，根据author_id
	user, err := mysql.GetUserByID(post.AuthorId)
	if err != nil {
		zap.L().Error("mysql.GetUserByID() failed",
			zap.Uint64("postID", post.AuthorId),
			zap.Error(err))
		return
	}
	// 3.查询社区详细信息，根据community_id
	community, err := mysql.GetCommunityByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityByID() failed",
			zap.Uint64("community_id", post.CommunityID),
			zap.Error(err))
		return
	}
	// 根据帖子id查询帖子的投票数
	voteNum, err := redis.GetPostVoteNum(postID)

	// 拼接帖子详情并返回
	data = &models.ApiPostDetail{
		Post:               post,
		CommunityDetailRes: community,
		AuthorName:         user.UserName,
		VoteNum:            voteNum,
	}
	return data, nil
}

// GetPostList 获取帖子列表
func GetPostList(page, size int64) ([]*models.ApiPostDetail, error) {
	// 1.获取帖子列表
	postList, err := mysql.GetPostList(page, size)
	if err != nil {
		zap.L().Error("mysql.GetPostList() failed")
		return nil, err
	}
	data := make([]*models.ApiPostDetail, 0, len(postList)) // init data
	// 2.遍历帖子列表，完善每个帖子的详细信息放入data
	for _, post := range postList {
		// 查询作者信息，根据author_id
		user, err := mysql.GetUserByID(post.AuthorId)
		if err != nil {
			zap.L().Error("mysql.GetUserByID() failed",
				zap.Uint64("postID", post.AuthorId),
				zap.Error(err))
			continue
		}
		// 查询社区详细信息，根据community_id
		community, err := mysql.GetCommunityByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityByID() failed",
				zap.Uint64("community_id", post.CommunityID),
				zap.Error(err))
			continue
		}
		// 拼接获得帖子详情
		postDetail := &models.ApiPostDetail{
			Post:               post,
			CommunityDetailRes: community,
			AuthorName:         user.UserName,
		}
		data = append(data, postDetail)
	}
	return data, nil
}

// GetPostList2 按发布时间/分数排序分页获取所有帖子列表
func GetPostList2(p *models.ParamPostList) (*models.ApiPostDetailRes, error) {
	var res models.ApiPostDetailRes
	// 1.从mysql获取所有帖子总数
	total, err := mysql.GetPostTotalCount()
	if err != nil {
		return nil, err
	}
	res.Page.Total = total

	// 2.根据排序规则(order)去redis查询帖子列表(ids)
	ids, err := redis.GetPostIDsInOrder(p)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetPostIDsInOrder(p) return 0 data")
		return &res, nil
	}
	zap.L().Debug("GetPostList2", zap.Any("ids: ", ids))

	// 3.查询ids中每篇帖子的赞成票数量
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return nil, err
	}

	// 4.根据ids去数据库查询帖子详细信息，并按传入的ids顺序返回结果
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		return nil, err
	}
	res.Page.Page = p.Page
	res.Page.Size = p.Size
	res.List = make([]*models.ApiPostDetail, 0, len(posts))
	// 5.拼接数据：将帖子的作者及分区信息查询出来填充到帖子中
	for idx, post := range posts {
		// 根据user_id查询作者信息
		user, err := mysql.GetUserByID(post.AuthorId)
		if err != nil {
			zap.L().Error("mysql.GetUserByID() failed",
				zap.Uint64("postID", post.AuthorId),
				zap.Error(err))
			user = nil
		}
		// 根据community_id查询社区详细信息
		community, err := mysql.GetCommunityByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityByID() failed",
				zap.Uint64("community_id", post.CommunityID),
				zap.Error(err))
			community = nil
		}
		// 拼接获得帖子详情信息
		postDetail := &models.ApiPostDetail{
			VoteNum:            voteData[idx],
			Post:               post,
			CommunityDetailRes: community,
			AuthorName:         user.UserName,
		}
		res.List = append(res.List, postDetail)
	}
	return &res, nil
}

// GetCommunityPostList 按发布时间/分数排序分页获取某社区的帖子列表
func GetCommunityPostList(p *models.ParamPostList) (*models.ApiPostDetailRes, error) {
	var res models.ApiPostDetailRes
	// 1.从mysql获取该社区下帖子总数
	total, err := mysql.GetCommunityPostTotalCount(p.CommunityID)
	if err != nil {
		return nil, err
	}
	res.Page.Total = total

	// 2.根据参数order去redis查询ids
	ids, err := redis.GetCommunityPostIDsInOrder(p)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetCommunityPostList(p) return 0 data")
		return &res, nil
	}
	zap.L().Debug("GetPostList2", zap.Any("ids", ids))

	// 3.查询ids中每篇帖子的赞成票数量
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return nil, err
	}

	// 4.根据ids去数据库查询帖子详细信息，并按传入的ids顺序返回结果
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		return nil, err
	}
	res.Page.Page = p.Page
	res.Page.Size = p.Size
	res.List = make([]*models.ApiPostDetail, 0, len(posts))

	// 5.拼接数据：将帖子的作者及分区信息查询出来填充到帖子中
	// 社区信息仅有一个，提前查询可以减少数据库的查询次数
	community, err := mysql.GetCommunityByID(p.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityByID() failed",
			zap.Uint64("community_id", p.CommunityID),
			zap.Error(err))
		community = nil
	}

	for idx, post := range posts {
		// 过滤掉不属于该社区的帖子
		if post.CommunityID != p.CommunityID {
			continue
		}
		// 根据作者id查询作者信息
		user, err := mysql.GetUserByID(post.AuthorId)
		if err != nil {
			zap.L().Error("mysql.GetUserByID() failed",
				zap.Uint64("postID", post.AuthorId),
				zap.Error(err))
			user = nil
		}
		// 拼接获得帖子详情信息
		postDetail := &models.ApiPostDetail{
			VoteNum:            voteData[idx],
			Post:               post,
			CommunityDetailRes: community,
			AuthorName:         user.UserName,
		}
		res.List = append(res.List, postDetail)
	}
	return &res, nil
}

// GetPostListNew 将两个查询帖子列表逻辑合二为一的函数
func GetPostListNew(p *models.ParamPostList) (data *models.ApiPostDetailRes, err error) {
	// 根据请求参数的不同,执行不同的业务逻辑
	if p.CommunityID == 0 {
		// 查所有帖子
		data, err = GetPostList2(p)
	} else {
		// 只查community_id社区下的帖子
		data, err = GetCommunityPostList(p)
	}
	if err != nil {
		zap.L().Error("GetPostListNew failed", zap.Error(err))
		return nil, err
	}
	return data, nil
}

// PostSearch 搜索业务-搜索帖子
func PostSearch(p *models.ParamPostList) (*models.ApiPostDetailRes, error) {
	var res models.ApiPostDetailRes
	// 根据搜索条件去mysql查询符合条件的帖子列表总数
	total, err := mysql.GetPostListTotalCount(p)
	if err != nil {
		return nil, err
	}
	res.Page.Total = total
	// 1、根据搜索条件去mysql分页查询符合条件的帖子列表
	posts, err := mysql.GetPostListByKeywords(p)
	if err != nil {
		return nil, err
	}
	// 查询出来的帖子总数可能为0
	if len(posts) == 0 {
		return &models.ApiPostDetailRes{}, nil
	}
	// 2、查询出来的帖子id列表传入到redis接口获取帖子的投票数
	ids := make([]string, 0, len(posts))
	for _, post := range posts {
		ids = append(ids, strconv.Itoa(int(post.PostID)))
	}
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return nil, err
	}
	res.Page.Size = p.Size
	res.Page.Page = p.Page
	// 3、拼接数据
	res.List = make([]*models.ApiPostDetail, 0, len(posts))
	for idx, post := range posts {
		// 根据作者id查询作者信息
		user, err := mysql.GetUserByID(post.AuthorId)
		if err != nil {
			zap.L().Error("mysql.GetUserByID() failed",
				zap.Uint64("postID", post.AuthorId),
				zap.Error(err))
			user = nil
		}
		// 根据社区id查询社区详细信息
		community, err := mysql.GetCommunityByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityByID() failed",
				zap.Uint64("community_id", post.CommunityID),
				zap.Error(err))
			community = nil
		}
		// 接口数据拼接
		postDetail := &models.ApiPostDetail{
			VoteNum:            voteData[idx],
			Post:               post,
			CommunityDetailRes: community,
			AuthorName:         user.UserName,
		}
		res.List = append(res.List, postDetail)
	}
	return &res, nil
}
