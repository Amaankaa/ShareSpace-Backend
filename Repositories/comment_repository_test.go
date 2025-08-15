package repositories_test

import (
	"context"
	"testing"
	"time"

	commentpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/comment"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

type CommentRepositoryTestSuite struct {
	suite.Suite
	mt   *mtest.T
	repo *repositories.CommentRepository
}

func TestCommentRepositoryTestSuite(t *testing.T) { suite.Run(t, new(CommentRepositoryTestSuite)) }

func (s *CommentRepositoryTestSuite) SetupSuite() {
	s.mt = mtest.New(s.T(), mtest.NewOptions().ClientType(mtest.Mock))
}

func (s *CommentRepositoryTestSuite) SetupTest() {
	s.mt.Run("test", func(mt *mtest.T) { s.repo = repositories.NewCommentRepository(mt.Coll) })
}

func (s *CommentRepositoryTestSuite) TestCreateComment_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewCommentRepository(mt.Coll)
		c := commentpkg.Comment{PostID: primitive.NewObjectID(), AuthorID: primitive.NewObjectID(), Content: "hi"}
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		res, err := s.repo.CreateComment(context.Background(), c)
		s.NoError(err)
		s.NotNil(res)
		s.Equal("hi", res.Content)
		s.NotZero(res.ID)
	})
}

func (s *CommentRepositoryTestSuite) TestGetCommentsByPost_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewCommentRepository(mt.Coll)
		postID := primitive.NewObjectID()
		authorID := primitive.NewObjectID()
		commentDoc := bson.D{{Key: "_id", Value: primitive.NewObjectID()}, {Key: "postId", Value: postID}, {Key: "authorId", Value: authorID}, {Key: "content", Value: "hello"}, {Key: "createdAt", Value: time.Now()}}
		// CountDocuments cursor with n=1
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.comments", mtest.FirstBatch, bson.D{{Key: "n", Value: 1}}))
		// Find returns cursor with comment
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.comments", mtest.FirstBatch, commentDoc))
		list, total, err := s.repo.GetCommentsByPost(context.Background(), postID, commentpkg.CommentPagination{Page: 1, PageSize: 20})
		s.NoError(err)
		s.Equal(int64(1), total)
		s.Len(list, 1)
		s.Equal("hello", list[0].Content)
	})
}

func (s *CommentRepositoryTestSuite) TestGetByID_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewCommentRepository(mt.Coll)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.comments", mtest.FirstBatch))
		res, err := s.repo.GetByID(context.Background(), primitive.NewObjectID())
		s.Error(err)
		s.Nil(res)
	})
}

func (s *CommentRepositoryTestSuite) TestUpdateComment_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewCommentRepository(mt.Coll)
		id := primitive.NewObjectID()
		postID := primitive.NewObjectID()
		authorID := primitive.NewObjectID()
		updated := bson.D{{Key: "_id", Value: id}, {Key: "postId", Value: postID}, {Key: "authorId", Value: authorID}, {Key: "content", Value: "updated"}, {Key: "updatedAt", Value: time.Now()}}
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "value", Value: updated}))
		res, err := s.repo.UpdateComment(context.Background(), id, "updated")
		s.NoError(err)
		s.NotNil(res)
		s.Equal("updated", res.Content)
	})
}

func (s *CommentRepositoryTestSuite) TestDeleteComment_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewCommentRepository(mt.Coll)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 0}))
		err := s.repo.DeleteComment(context.Background(), primitive.NewObjectID())
		s.Error(err)
	})
}
