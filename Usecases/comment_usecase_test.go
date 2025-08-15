package usecases_test

import (
	"context"
	"errors"
	"testing"

	commentpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/comment"
	postpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/post"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentUsecaseTestSuite struct {
	suite.Suite
	commentRepo *mocks.ICommentRepository
	postRepo    *mocks.PostRepository
	userRepo    *mocks.IUserRepository
	uc          *usecases.CommentUsecase
}

func TestCommentUsecaseTestSuite(t *testing.T) { suite.Run(t, new(CommentUsecaseTestSuite)) }

func (s *CommentUsecaseTestSuite) SetupTest() {
	s.commentRepo = &mocks.ICommentRepository{}
	s.postRepo = mocks.NewPostRepository(s.T())
	s.userRepo = &mocks.IUserRepository{}
	s.uc = usecases.NewCommentUsecase(s.commentRepo, s.postRepo, s.userRepo)
}

func (s *CommentUsecaseTestSuite) TearDownTest() {
	s.commentRepo.AssertExpectations(s.T())
	s.postRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

func (s *CommentUsecaseTestSuite) TestCreateComment_Success() {
	postID := primitive.NewObjectID()
	uid := primitive.NewObjectID()
	s.postRepo.On("GetPostByID", mock.Anything, postID).Return(&postpkg.Post{ID: postID, AuthorID: primitive.NewObjectID()}, nil).Once()
	s.commentRepo.On("CreateComment", mock.Anything, mock.AnythingOfType("comment.Comment")).Return(&commentpkg.Comment{ID: primitive.NewObjectID(), PostID: postID, AuthorID: uid, Content: "hi"}, nil).Once()
	s.postRepo.On("UpdateCommentsCount", mock.Anything, postID, 1).Return(nil).Maybe()
	s.userRepo.On("FindByID", mock.Anything, uid.Hex()).Return(userpkg.User{ID: uid, DisplayName: "Tester", ProfilePicture: ""}, nil).Once()
	resp, err := s.uc.CreateComment(context.Background(), postID, commentpkg.CreateCommentRequest{Content: "hi"}, uid)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal("hi", resp.Content)
}

func (s *CommentUsecaseTestSuite) TestUpdateComment_Unauthorized() {
	cid := primitive.NewObjectID()
	uid := primitive.NewObjectID()
	s.commentRepo.On("GetByID", mock.Anything, cid).Return(&commentpkg.Comment{ID: cid, AuthorID: primitive.NewObjectID(), PostID: primitive.NewObjectID(), Content: "x"}, nil).Once()
	_, err := s.uc.UpdateComment(context.Background(), cid, commentpkg.UpdateCommentRequest{Content: "y"}, uid)
	s.Error(err)
	s.Contains(err.Error(), "unauthorized")
}

func (s *CommentUsecaseTestSuite) TestDeleteComment_Success() {
	cid := primitive.NewObjectID()
	uid := primitive.NewObjectID()
	postID := primitive.NewObjectID()
	s.commentRepo.On("GetByID", mock.Anything, cid).Return(&commentpkg.Comment{ID: cid, AuthorID: uid, PostID: postID, Content: "x"}, nil).Once()
	s.commentRepo.On("DeleteComment", mock.Anything, cid).Return(nil).Once()
	s.postRepo.On("UpdateCommentsCount", mock.Anything, postID, -1).Return(nil).Maybe()
	err := s.uc.DeleteComment(context.Background(), cid, uid)
	s.NoError(err)
}

func (s *CommentUsecaseTestSuite) TestCreateComment_EmptyContent() {
	postID := primitive.NewObjectID()
	uid := primitive.NewObjectID()
	_, err := s.uc.CreateComment(context.Background(), postID, commentpkg.CreateCommentRequest{Content: "  "}, uid)
	s.Error(err)
}

func (s *CommentUsecaseTestSuite) TestGetComments_PostNotFound() {
	postID := primitive.NewObjectID()
	s.postRepo.On("GetPostByID", mock.Anything, postID).Return(nil, errors.New("post not found")).Once()
	_, err := s.uc.GetComments(context.Background(), postID, commentpkg.CommentPagination{Page: 1, PageSize: 20})
	s.Error(err)
}
