package mysql

import (
	"bluebell_backend/models"

	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

// CreateComment
func CreateComment(comment *models.Comment) (err error) {
	sqlStr := `insert into comment(
	comment_id, content, post_id, author_id, parent_id)
	values(?,?,?,?,?)`
	_, err = db.Exec(sqlStr, comment.CommentID, comment.Content, comment.PostID,
		comment.AuthorID, comment.ParentID)
	if err != nil {
		zap.L().Error("insert comment failed", zap.Error(err))
		err = ErrorInsertFailed
		return
	}
	return
}

func GetCommentListByIDs(ids []string) (commentList []*models.Comment, err error) {
	sqlStr := `select comment_id, content, post_id, author_id, parent_id, create_time
	from comment
	where comment_id in (?)`
	// 使用 sqlx.In 动态生成带有占位符的SQL查询语句，并将参数绑定到查询中
	query, args, err := sqlx.In(sqlStr, ids)
	if err != nil {
		return
	}
	// 重新绑定查询语句，确保占位符正确
	query = db.Rebind(query)
	err = db.Select(&commentList, query, args...)
	return
}
