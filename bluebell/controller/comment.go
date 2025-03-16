package controller

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CommentHandler 创建评论
func CommentHandler(c *gin.Context) {
	// 1.接收参数
	var comment models.Comment
	if err := c.BindJSON(&comment); err != nil {
		fmt.Println(err)
		ResponseError(c, CodeInvalidParams)
		return
	}

	// 雪花算法生成评论ID
	commentID, err := snowflake.GetID()
	if err != nil {
		zap.L().Error("snowflake.GetID() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 获取作者ID，从请求上下文的userID获取
	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("GetCurrentUserID() failed", zap.Error(err))
		ResponseError(c, CodeNotLogin)
		return
	}
	comment.CommentID = commentID
	comment.AuthorID = userID

	// 2.在数据库中插入评论
	if err := mysql.CreateComment(&comment); err != nil {
		zap.L().Error("mysql.CreateComment(&comment) failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

// CommentListHandler 评论列表
func CommentListHandler(c *gin.Context) {
	// https://example.com/api/v1/comments?ids=1&ids=2&ids=3
	// 1.接收参数 从url查询字符串中获取评论列表
	ids, ok := c.GetQueryArray("ids")
	if !ok {
		ResponseError(c, CodeInvalidParams)
		return
	}
	// 2.从数据库中获取每条评论的详细信息
	posts, err := mysql.GetCommentListByIDs(ids)
	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, posts)
}
