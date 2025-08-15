package repositories_test

import (
	"context"
	"testing"
	"time"

	postpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/post"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

type PostRepositoryTestSuite struct {
	suite.Suite
	mt   *mtest.T
	repo *repositories.PostRepository
}

func TestPostRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PostRepositoryTestSuite))
}

func (s *PostRepositoryTestSuite) SetupSuite() {
	s.mt = mtest.New(s.T(), mtest.NewOptions().ClientType(mtest.Mock))
}

func (s *PostRepositoryTestSuite) TearDownSuite() {
	// mtest.T doesn't have a Close method
}

func (s *PostRepositoryTestSuite) SetupTest() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewPostRepository(mt.Coll)
	})
}

// Test CreatePost
func (s *PostRepositoryTestSuite) TestCreatePost_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		post := postpkg.Post{
			AuthorID: primitive.NewObjectID(),
			Title:    "Test Post",
			Content:  "Test content",
			Category: "Academic Struggles",
			Tags:     []string{"test"},
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Act
		result, err := s.repo.CreatePost(context.Background(), post)

		// Assert
		s.NoError(err)
		s.NotNil(result)
		s.Equal("Test Post", result.Title)
		s.Equal("Test content", result.Content)
		s.Equal("Academic Struggles", result.Category)
		s.Equal(postpkg.PostStatusActive, result.Status)
		s.Equal(0, result.LikesCount)
		s.Equal(0, result.CommentsCount)
		s.Equal(0, result.ViewsCount)
		s.NotZero(result.ID)
		s.NotZero(result.CreatedAt)
		s.NotZero(result.UpdatedAt)
	})
}

// Test GetPostByID
func (s *PostRepositoryTestSuite) TestGetPostByID_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		authorID := primitive.NewObjectID()

		expectedPost := bson.D{
			{Key: "_id", Value: postID},
			{Key: "authorId", Value: authorID},
			{Key: "title", Value: "Test Post"},
			{Key: "content", Value: "Test content"},
			{Key: "category", Value: "Academic Struggles"},
			{Key: "status", Value: postpkg.PostStatusActive},
			{Key: "likesCount", Value: 5},
			{Key: "commentsCount", Value: 3},
			{Key: "viewsCount", Value: 100},
			{Key: "createdAt", Value: time.Now()},
			{Key: "updatedAt", Value: time.Now()},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "blog_db.posts", mtest.FirstBatch, expectedPost))

		// Act
		result, err := s.repo.GetPostByID(context.Background(), postID)

		// Assert
		s.NoError(err)
		s.NotNil(result)
		s.Equal(postID, result.ID)
		s.Equal(authorID, result.AuthorID)
		s.Equal("Test Post", result.Title)
		s.Equal("Test content", result.Content)
		s.Equal("Academic Struggles", result.Category)
		s.Equal(5, result.LikesCount)
		s.Equal(3, result.CommentsCount)
		s.Equal(100, result.ViewsCount)
	})
}

func (s *PostRepositoryTestSuite) TestGetPostByID_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.posts", mtest.FirstBatch))

		// Act
		result, err := s.repo.GetPostByID(context.Background(), postID)

		// Assert
		s.Error(err)
		s.Nil(result)
		s.Contains(err.Error(), "post not found")
	})
}

// Test UpdatePost
func (s *PostRepositoryTestSuite) TestUpdatePost_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		authorID := primitive.NewObjectID()

		updates := postpkg.Post{
			Title:   "Updated Title",
			Content: "Updated content",
		}

		updatedPost := bson.D{
			{Key: "_id", Value: postID},
			{Key: "authorId", Value: authorID},
			{Key: "title", Value: "Updated Title"},
			{Key: "content", Value: "Updated content"},
			{Key: "category", Value: "Academic Struggles"},
			{Key: "status", Value: postpkg.PostStatusActive},
			{Key: "updatedAt", Value: time.Now()},
		}

	// FindOneAndUpdate returns a document in the 'value' field, not a cursor
	mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "value", Value: updatedPost}))

		// Act
		result, err := s.repo.UpdatePost(context.Background(), postID, updates)

		// Assert
		s.NoError(err)
		s.NotNil(result)
		s.Equal("Updated Title", result.Title)
		s.Equal("Updated content", result.Content)
	})
}

func (s *PostRepositoryTestSuite) TestUpdatePost_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		updates := postpkg.Post{Title: "Updated Title"}

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.posts", mtest.FirstBatch))

		// Act
		result, err := s.repo.UpdatePost(context.Background(), postID, updates)

		// Assert
		s.Error(err)
		s.Nil(result)
		s.Contains(err.Error(), "post not found")
	})
}

// Test DeletePost
func (s *PostRepositoryTestSuite) TestDeletePost_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		// UpdateOne expects a write success response; set n (matched) to 1
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		))

		// Act
		err := s.repo.DeletePost(context.Background(), postID)

		// Assert
		s.NoError(err)
	})
}

func (s *PostRepositoryTestSuite) TestDeletePost_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		// UpdateOne matched 0 documents
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
		))

		// Act
		err := s.repo.DeletePost(context.Background(), postID)

		// Assert
		s.Error(err)
		s.Contains(err.Error(), "post not found")
	})
}

// Test GetPosts
func (s *PostRepositoryTestSuite) TestGetPosts_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID1 := primitive.NewObjectID()
		postID2 := primitive.NewObjectID()
		authorID := primitive.NewObjectID()

		post1 := bson.D{
			{Key: "_id", Value: postID1},
			{Key: "authorId", Value: authorID},
			{Key: "title", Value: "Post 1"},
			{Key: "content", Value: "Content 1"},
			{Key: "category", Value: "Academic Struggles"},
			{Key: "status", Value: postpkg.PostStatusActive},
			{Key: "createdAt", Value: time.Now()},
		}

		post2 := bson.D{
			{Key: "_id", Value: postID2},
			{Key: "authorId", Value: authorID},
			{Key: "title", Value: "Post 2"},
			{Key: "content", Value: "Content 2"},
			{Key: "category", Value: "Academic Struggles"},
			{Key: "status", Value: postpkg.PostStatusActive},
			{Key: "createdAt", Value: time.Now()},
		}

	// Mock count response for CountDocuments (aggregate style with cursor)
	mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.posts", mtest.FirstBatch, bson.D{{Key: "n", Value: 2}}))
	// Mock find response with single batch (cursorID 0 means no getMore)
	mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.posts", mtest.FirstBatch, post1, post2))

		filter := postpkg.PostFilter{Category: "Academic Struggles"}
		pagination := postpkg.PostPagination{Page: 1, PageSize: 10}
	
		// Act
		posts, total, err := s.repo.GetPosts(context.Background(), filter, pagination)

		// Assert
		s.NoError(err)
		s.Equal(int64(2), total)
		s.Len(posts, 2)
		s.Equal("Post 1", posts[0].Title)
		s.Equal("Post 2", posts[1].Title)
	})
}

// Test LikePost
func (s *PostRepositoryTestSuite) TestLikePost_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))

		// Act
		err := s.repo.LikePost(context.Background(), postID, userID)

		// Assert
		s.NoError(err)
	})
}

func (s *PostRepositoryTestSuite) TestLikePost_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 0}))

		// Act
		err := s.repo.LikePost(context.Background(), postID, userID)

		// Assert
		s.Error(err)
		s.Contains(err.Error(), "post not found")
	})
}

// Test UnlikePost
func (s *PostRepositoryTestSuite) TestUnlikePost_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))

		// Act
		err := s.repo.UnlikePost(context.Background(), postID, userID)

		// Assert
		s.NoError(err)
	})
}

// Test IsPostLikedByUser
func (s *PostRepositoryTestSuite) TestIsPostLikedByUser_True() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

	// CountDocuments returns a cursor with a doc containing field 'n'
	mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.posts", mtest.FirstBatch, bson.D{{Key: "n", Value: 1}}))

		// Act
		isLiked, err := s.repo.IsPostLikedByUser(context.Background(), postID, userID)

		// Assert
		s.NoError(err)
		s.True(isLiked)
	})
}

func (s *PostRepositoryTestSuite) TestIsPostLikedByUser_False() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

	// CountDocuments returns a cursor with a doc containing field 'n'
	mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.posts", mtest.FirstBatch, bson.D{{Key: "n", Value: 0}}))

		// Act
		isLiked, err := s.repo.IsPostLikedByUser(context.Background(), postID, userID)

		// Assert
		s.NoError(err)
		s.False(isLiked)
	})
}

// Test IncrementViewCount
func (s *PostRepositoryTestSuite) TestIncrementViewCount_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Act
		err := s.repo.IncrementViewCount(context.Background(), postID)

		// Assert
		s.NoError(err)
	})
}

// Test UpdateCommentsCount
func (s *PostRepositoryTestSuite) TestUpdateCommentsCount_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Act
		err := s.repo.UpdateCommentsCount(context.Background(), postID, 1)

		// Assert
		s.NoError(err)
	})
}

// Test SearchPosts
func (s *PostRepositoryTestSuite) TestSearchPosts_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		// Arrange
		s.repo = repositories.NewPostRepository(mt.Coll)

		postID := primitive.NewObjectID()
		authorID := primitive.NewObjectID()

		post := bson.D{
			{Key: "_id", Value: postID},
			{Key: "authorId", Value: authorID},
			{Key: "title", Value: "Test Search Post"},
			{Key: "content", Value: "This contains the search term"},
			{Key: "category", Value: "Academic Struggles"},
			{Key: "status", Value: postpkg.PostStatusActive},
		}

	// Mock count response for CountDocuments (cursor)
	mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.posts", mtest.FirstBatch, bson.D{{Key: "n", Value: 1}}))
	// Mock find response with single batch
	mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.posts", mtest.FirstBatch, post))

		filter := postpkg.PostFilter{}
		pagination := postpkg.PostPagination{Page: 1, PageSize: 10}

		// Act
		posts, total, err := s.repo.SearchPosts(context.Background(), "search term", filter, pagination)

		// Assert
		s.NoError(err)
		s.Equal(int64(1), total)
		s.Len(posts, 1)
		s.Equal("Test Search Post", posts[0].Title)
	})
}
