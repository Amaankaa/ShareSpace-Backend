package repositories_test

import (
	"context"
	"testing"
	"time"

	resourcepkg "github.com/Amaankaa/Blog-Starter-Project/Domain/resource"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

type ResourceRepositoryTestSuite struct {
	suite.Suite
	mt   *mtest.T
	repo *repositories.ResourceRepository
}

func TestResourceRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceRepositoryTestSuite))
}

func (s *ResourceRepositoryTestSuite) SetupSuite() {
	s.mt = mtest.New(s.T(), mtest.NewOptions().ClientType(mtest.Mock))
}

func (s *ResourceRepositoryTestSuite) SetupTest() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
	})
}

// CreateResource
func (s *ResourceRepositoryTestSuite) TestCreateResource_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)

		res := resourcepkg.Resource{
			CreatorID:   primitive.NewObjectID(),
			Title:       "Test Resource",
			Description: "A helpful resource",
			Content:     "Body...",
			Type:        "guide",
			Category:    "Academic Success",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		out, err := s.repo.CreateResource(context.Background(), res)
		s.NoError(err)
		s.NotNil(out)
		s.Equal("Test Resource", out.Title)
		s.Equal(resourcepkg.ResourceStatusActive, out.Status)
		s.NotZero(out.ID)
	})
}

// GetResourceByID
func (s *ResourceRepositoryTestSuite) TestGetResourceByID_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)

		id := primitive.NewObjectID()
		creator := primitive.NewObjectID()
		doc := bson.D{
			{Key: "_id", Value: id},
			{Key: "creatorId", Value: creator},
			{Key: "title", Value: "Test Resource"},
			{Key: "description", Value: "A helpful resource"},
			{Key: "content", Value: "Body..."},
			{Key: "type", Value: "guide"},
			{Key: "category", Value: "Academic Success"},
			{Key: "status", Value: resourcepkg.ResourceStatusActive},
			{Key: "createdAt", Value: time.Now()},
			{Key: "updatedAt", Value: time.Now()},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "blog_db.resources", mtest.FirstBatch, doc))

		out, err := s.repo.GetResourceByID(context.Background(), id)
		s.NoError(err)
		s.Equal(id, out.ID)
		s.Equal("Test Resource", out.Title)
	})
}

func (s *ResourceRepositoryTestSuite) TestGetResourceByID_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch))
		out, err := s.repo.GetResourceByID(context.Background(), id)
		s.Error(err)
		s.Nil(out)
		s.Contains(err.Error(), "resource not found")
	})
}

// UpdateResource
func (s *ResourceRepositoryTestSuite) TestUpdateResource_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		updates := resourcepkg.Resource{Title: "Updated", Content: "New body"}
		updated := bson.D{
			{Key: "_id", Value: id},
			{Key: "title", Value: "Updated"},
			{Key: "content", Value: "New body"},
			{Key: "status", Value: resourcepkg.ResourceStatusActive},
			{Key: "updatedAt", Value: time.Now()},
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "value", Value: updated}))
		out, err := s.repo.UpdateResource(context.Background(), id, updates)
		s.NoError(err)
		s.Equal("Updated", out.Title)
	})
}

func (s *ResourceRepositoryTestSuite) TestUpdateResource_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		updates := resourcepkg.Resource{Title: "Updated"}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch))
		out, err := s.repo.UpdateResource(context.Background(), id, updates)
		s.Error(err)
		s.Nil(out)
		s.Contains(err.Error(), "resource not found")
	})
}

// DeleteResource
func (s *ResourceRepositoryTestSuite) TestDeleteResource_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}))
		err := s.repo.DeleteResource(context.Background(), id)
		s.NoError(err)
	})
}

func (s *ResourceRepositoryTestSuite) TestDeleteResource_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 0}, bson.E{Key: "nModified", Value: 0}))
		err := s.repo.DeleteResource(context.Background(), id)
		s.Error(err)
		s.Contains(err.Error(), "resource not found")
	})
}

// GetResources (count + find)
func (s *ResourceRepositoryTestSuite) TestGetResources_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()
		creator := primitive.NewObjectID()
		res1 := bson.D{{Key: "_id", Value: id1}, {Key: "creatorId", Value: creator}, {Key: "title", Value: "R1"}, {Key: "status", Value: resourcepkg.ResourceStatusActive}, {Key: "createdAt", Value: time.Now()}}
		res2 := bson.D{{Key: "_id", Value: id2}, {Key: "creatorId", Value: creator}, {Key: "title", Value: "R2"}, {Key: "status", Value: resourcepkg.ResourceStatusActive}, {Key: "createdAt", Value: time.Now()}}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, bson.D{{Key: "n", Value: 2}}))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, res1, res2))
		filter := resourcepkg.ResourceFilter{Category: "Academic Success"}
		pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 10}
		items, total, err := s.repo.GetResources(context.Background(), filter, pg)
		s.NoError(err)
		s.Equal(int64(2), total)
		s.Len(items, 2)
		s.Equal("R1", items[0].Title)
		s.Equal("R2", items[1].Title)
	})
}

// LikeResource
func (s *ResourceRepositoryTestSuite) TestLikeResource_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.LikeResource(context.Background(), id, user))
	})
}

func (s *ResourceRepositoryTestSuite) TestLikeResource_NotFound() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 0}))
		err := s.repo.LikeResource(context.Background(), id, user)
		s.Error(err)
	})
}

// UnlikeResource
func (s *ResourceRepositoryTestSuite) TestUnlikeResource_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.UnlikeResource(context.Background(), id, user))
	})
}

// BookmarkResource / UnbookmarkResource
func (s *ResourceRepositoryTestSuite) TestBookmarkResource_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.BookmarkResource(context.Background(), id, user))
	})
}

func (s *ResourceRepositoryTestSuite) TestUnbookmarkResource_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.UnbookmarkResource(context.Background(), id, user))
	})
}

// IsResourceLikedByUser
func (s *ResourceRepositoryTestSuite) TestIsResourceLikedByUser_True() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, bson.D{{Key: "n", Value: 1}}))
		ok, err := s.repo.IsResourceLikedByUser(context.Background(), id, user)
		s.NoError(err)
		s.True(ok)
	})
}

func (s *ResourceRepositoryTestSuite) TestIsResourceLikedByUser_False() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, bson.D{{Key: "n", Value: 0}}))
		ok, err := s.repo.IsResourceLikedByUser(context.Background(), id, user)
		s.NoError(err)
		s.False(ok)
	})
}

// IsResourceBookmarkedByUser
func (s *ResourceRepositoryTestSuite) TestIsResourceBookmarkedByUser_True() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, bson.D{{Key: "n", Value: 1}}))
		ok, err := s.repo.IsResourceBookmarkedByUser(context.Background(), id, user)
		s.NoError(err)
		s.True(ok)
	})
}

func (s *ResourceRepositoryTestSuite) TestIsResourceBookmarkedByUser_False() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		user := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, bson.D{{Key: "n", Value: 0}}))
		ok, err := s.repo.IsResourceBookmarkedByUser(context.Background(), id, user)
		s.NoError(err)
		s.False(ok)
	})
}

// Increment counters
func (s *ResourceRepositoryTestSuite) TestIncrementViewCount_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		s.NoError(s.repo.IncrementViewCount(context.Background(), id))
	})
}

func (s *ResourceRepositoryTestSuite) TestIncrementShareCount_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		s.NoError(s.repo.IncrementShareCount(context.Background(), id))
	})
}

// SearchResources
func (s *ResourceRepositoryTestSuite) TestSearchResources_Success() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		doc := bson.D{{Key: "_id", Value: id}, {Key: "title", Value: "Search Hit"}, {Key: "status", Value: resourcepkg.ResourceStatusActive}}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, bson.D{{Key: "n", Value: 1}}))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, doc))
		filter := resourcepkg.ResourceFilter{}
		pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 10}
		items, total, err := s.repo.SearchResources(context.Background(), "Search", filter, pg)
		s.NoError(err)
		s.Equal(int64(1), total)
		s.Len(items, 1)
	})
}

// Verification
func (s *ResourceRepositoryTestSuite) TestVerifyAndUnverifyResource() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		verifier := primitive.NewObjectID()
		// Verify
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.VerifyResource(context.Background(), id, verifier))
		// Unverify
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.UnverifyResource(context.Background(), id))
	})
}

// Moderation
func (s *ResourceRepositoryTestSuite) TestReportHideUnhide() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.ReportResource(context.Background(), id))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.HideResource(context.Background(), id))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		s.NoError(s.repo.UnhideResource(context.Background(), id))
	})
}

// Deadlines
func (s *ResourceRepositoryTestSuite) TestUpcomingAndExpiredDeadlines() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		doc := bson.D{{Key: "_id", Value: id}, {Key: "title", Value: "Opportunity"}, {Key: "status", Value: resourcepkg.ResourceStatusActive}, {Key: "deadline", Value: time.Now().Add(48 * time.Hour)}}
		pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 10}
		// Upcoming: count + find
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, bson.D{{Key: "n", Value: 1}}))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, doc))
		items, total, err := s.repo.GetResourcesWithUpcomingDeadlines(context.Background(), 7, pg)
		s.NoError(err)
		s.Equal(int64(1), total)
		s.Len(items, 1)
		// Expired: count + find
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, bson.D{{Key: "n", Value: 0}}))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch))
		items2, total2, err2 := s.repo.GetExpiredResources(context.Background(), pg)
		s.NoError(err2)
		s.Equal(int64(0), total2)
		s.Len(items2, 0)
	})
}

// Popular / TopRated / Trending
func (s *ResourceRepositoryTestSuite) TestPopularTopRatedTrending() {
	s.mt.Run("test", func(mt *mtest.T) {
		s.repo = repositories.NewResourceRepository(mt.Coll)
		id := primitive.NewObjectID()
		doc := bson.D{{Key: "_id", Value: id}, {Key: "title", Value: "Res"}, {Key: "status", Value: resourcepkg.ResourceStatusActive}}
		// Popular
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, doc))
		items, err := s.repo.GetPopularResources(context.Background(), 5, "week")
		s.NoError(err)
		s.Len(items, 1)
		// Trending
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, doc))
		items2, err2 := s.repo.GetTrendingResources(context.Background(), 5)
		s.NoError(err2)
		s.Len(items2, 1)
		// TopRated
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "blog_db.resources", mtest.FirstBatch, doc))
		items3, err3 := s.repo.GetTopRatedResources(context.Background(), 5, "")
		s.NoError(err3)
		s.Len(items3, 1)
	})
}
