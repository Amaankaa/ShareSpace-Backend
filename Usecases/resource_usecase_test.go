package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	resourcepkg "github.com/Amaankaa/Blog-Starter-Project/Domain/resource"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResourceUsecaseTestSuite struct {
	suite.Suite
	ctx       context.Context
	mockRepo  *mocks.ResourceRepository
	mockUsers *mocks.IUserRepository
	usecase   *usecases.ResourceUsecase
}

func TestResourceUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceUsecaseTestSuite))
}

func (s *ResourceUsecaseTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockRepo = new(mocks.ResourceRepository)
	s.mockUsers = new(mocks.IUserRepository)
	s.usecase = usecases.NewResourceUsecase(s.mockRepo, s.mockUsers)
}

func (s *ResourceUsecaseTestSuite) TearDownTest() {
	s.mockRepo.AssertExpectations(s.T())
	s.mockUsers.AssertExpectations(s.T())
}

func (s *ResourceUsecaseTestSuite) TestCreateResource_Success() {
	creatorID := primitive.NewObjectID()
	req := resourcepkg.CreateResourceRequest{
		Title:       "Test Resource",
		Description: "A helpful resource",
		Content:     "Content",
		Type:        "guide",
		Category:    "Academic Success",
		Tags:        []string{"Go", "Testing"},
	}

	creator := userpkg.User{ID: creatorID, DisplayName: "Alice", IsMentor: true, IsVerified: true}
	s.mockUsers.On("FindByID", mock.Anything, creatorID.Hex()).Return(creator, nil)

	created := &resourcepkg.Resource{ID: primitive.NewObjectID(), CreatorID: creatorID, Title: req.Title, Description: req.Description, Content: req.Content, Type: req.Type, Category: req.Category, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.mockRepo.On("CreateResource", mock.Anything, mock.AnythingOfType("resourcepkg.Resource")).Return(created, nil)

	res, err := s.usecase.CreateResource(s.ctx, req, creatorID)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(req.Title, res.Title)
	s.Equal(creatorID, res.Creator.ID)
}

func (s *ResourceUsecaseTestSuite) TestCreateResource_InvalidType() {
	creatorID := primitive.NewObjectID()
	// No mocks needed because validation fails before repo calls
	_, err := s.usecase.CreateResource(s.ctx, resourcepkg.CreateResourceRequest{Title: "T", Description: "D", Content: "C", Type: "bad", Category: "Academic Success"}, creatorID)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestGetResource_Success_WithViewerFlags() {
	id := primitive.NewObjectID()
	creatorID := primitive.NewObjectID()
	viewerID := primitive.NewObjectID()
	res := &resourcepkg.Resource{ID: id, CreatorID: creatorID, Title: "R", Description: "D", Content: "C", Type: "guide", Category: "Academic Success"}
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(res, nil)
	s.mockUsers.On("FindByID", mock.Anything, creatorID.Hex()).Return(userpkg.User{ID: creatorID, DisplayName: "Alice"}, nil)
	s.mockRepo.On("IncrementViewCount", mock.Anything, id).Return(nil)
	s.mockRepo.On("IsResourceLikedByUser", mock.Anything, id, viewerID).Return(true, nil)
	s.mockRepo.On("IsResourceBookmarkedByUser", mock.Anything, id, viewerID).Return(false, nil)

	resp, err := s.usecase.GetResource(s.ctx, id, &viewerID)
	s.NoError(err)
	s.NotNil(resp)
	s.True(resp.IsLikedByUser)
	s.False(resp.IsBookmarkedByUser)
}

func (s *ResourceUsecaseTestSuite) TestGetResource_NotFound() {
	id := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(nil, errors.New("not found"))
	_, err := s.usecase.GetResource(s.ctx, id, nil)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestUpdateResource_Unauthorized() {
	id := primitive.NewObjectID()
	creatorID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	existing := &resourcepkg.Resource{ID: id, CreatorID: creatorID, Type: "guide", Category: "Academic Success"}
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(existing, nil)

	_, err := s.usecase.UpdateResource(s.ctx, id, resourcepkg.UpdateResourceRequest{Title: "X"}, otherID)
	s.Error(err)
	s.Contains(err.Error(), "unauthorized")
}

func (s *ResourceUsecaseTestSuite) TestUpdateResource_Success() {
	id := primitive.NewObjectID()
	creatorID := primitive.NewObjectID()
	existing := &resourcepkg.Resource{ID: id, CreatorID: creatorID, Type: "guide", Category: "Academic Success"}
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(existing, nil)
	updated := &resourcepkg.Resource{ID: id, CreatorID: creatorID, Title: "X"}
	s.mockRepo.On("UpdateResource", mock.Anything, id, mock.AnythingOfType("resourcepkg.Resource")).Return(updated, nil)
	s.mockUsers.On("FindByID", mock.Anything, creatorID.Hex()).Return(userpkg.User{ID: creatorID}, nil)
	resp, err := s.usecase.UpdateResource(s.ctx, id, resourcepkg.UpdateResourceRequest{Title: "X"}, creatorID)
	s.NoError(err)
	s.Equal("X", resp.Title)
}

func (s *ResourceUsecaseTestSuite) TestLikeResource_AlreadyLiked() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id, CreatorID: userID}, nil)
	s.mockRepo.On("IsResourceLikedByUser", mock.Anything, id, userID).Return(true, nil)

	err := s.usecase.LikeResource(s.ctx, id, userID)
	s.Error(err)
	s.Contains(err.Error(), "already liked")
}

func (s *ResourceUsecaseTestSuite) TestLikeResource_Success() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("IsResourceLikedByUser", mock.Anything, id, userID).Return(false, nil)
	s.mockRepo.On("LikeResource", mock.Anything, id, userID).Return(nil)
	err := s.usecase.LikeResource(s.ctx, id, userID)
	s.NoError(err)
}

func (s *ResourceUsecaseTestSuite) TestUnlikeResource_Success() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("IsResourceLikedByUser", mock.Anything, id, userID).Return(true, nil)
	s.mockRepo.On("UnlikeResource", mock.Anything, id, userID).Return(nil)
	err := s.usecase.UnlikeResource(s.ctx, id, userID)
	s.NoError(err)
}

func (s *ResourceUsecaseTestSuite) TestUnlikeResource_NotLiked() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("IsResourceLikedByUser", mock.Anything, id, userID).Return(false, nil)
	err := s.usecase.UnlikeResource(s.ctx, id, userID)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestBookmarkResource_Success() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("IsResourceBookmarkedByUser", mock.Anything, id, userID).Return(false, nil)
	s.mockRepo.On("BookmarkResource", mock.Anything, id, userID).Return(nil)
	err := s.usecase.BookmarkResource(s.ctx, id, userID)
	s.NoError(err)
}

func (s *ResourceUsecaseTestSuite) TestBookmarkResource_AlreadyBookmarked() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("IsResourceBookmarkedByUser", mock.Anything, id, userID).Return(true, nil)
	err := s.usecase.BookmarkResource(s.ctx, id, userID)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestUnbookmarkResource_Success() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("IsResourceBookmarkedByUser", mock.Anything, id, userID).Return(true, nil)
	s.mockRepo.On("UnbookmarkResource", mock.Anything, id, userID).Return(nil)
	err := s.usecase.UnbookmarkResource(s.ctx, id, userID)
	s.NoError(err)
}

func (s *ResourceUsecaseTestSuite) TestUnbookmarkResource_NotBookmarked() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("IsResourceBookmarkedByUser", mock.Anything, id, userID).Return(false, nil)
	err := s.usecase.UnbookmarkResource(s.ctx, id, userID)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestRateResource_Success() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("RateResource", mock.Anything, id, userID, 4.0).Return(nil)
	err := s.usecase.RateResource(s.ctx, id, userID, 4.0)
	s.NoError(err)
}

func (s *ResourceUsecaseTestSuite) TestRateResource_InvalidValue() {
	id := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	_, _ = id, userID
	err := s.usecase.RateResource(s.ctx, id, userID, 7.0)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestSearchResources_Success() {
	pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 20}
	items := []resourcepkg.Resource{{ID: primitive.NewObjectID(), CreatorID: primitive.NewObjectID()}}
	s.mockRepo.On("SearchResources", mock.Anything, "go", mock.AnythingOfType("resourcepkg.ResourceFilter"), mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(items, int64(1), nil)
	s.mockUsers.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(userpkg.User{ID: items[0].CreatorID}, nil)
	s.mockRepo.On("IsResourceLikedByUser", mock.Anything, items[0].ID, mock.AnythingOfType("primitive.ObjectID")).Return(false, nil).Maybe()
	s.mockRepo.On("IsResourceBookmarkedByUser", mock.Anything, items[0].ID, mock.AnythingOfType("primitive.ObjectID")).Return(false, nil).Maybe()
	resp, err := s.usecase.SearchResources(s.ctx, "go", resourcepkg.ResourceFilter{}, pg, nil)
	s.NoError(err)
	s.Equal(int64(1), resp.Total)
}

func (s *ResourceUsecaseTestSuite) TestGetPopularResources_Success() {
	items := []resourcepkg.Resource{{ID: primitive.NewObjectID(), CreatorID: primitive.NewObjectID()}}
	s.mockRepo.On("GetPopularResources", mock.Anything, 20, "week").Return(items, nil)
	s.mockUsers.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(userpkg.User{ID: items[0].CreatorID}, nil)
	resp, err := s.usecase.GetPopularResources(s.ctx, 20, "week", nil)
	s.NoError(err)
	s.Equal(int64(1), resp.Total)
}

func (s *ResourceUsecaseTestSuite) TestGetPopularResources_Error() {
	s.mockRepo.On("GetPopularResources", mock.Anything, 20, "week").Return(nil, errors.New("err"))
	_, err := s.usecase.GetPopularResources(s.ctx, 20, "week", nil)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestGetTrendingResources_Success() {
	items := []resourcepkg.Resource{{ID: primitive.NewObjectID(), CreatorID: primitive.NewObjectID()}}
	s.mockRepo.On("GetTrendingResources", mock.Anything, 20).Return(items, nil)
	s.mockUsers.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(userpkg.User{ID: items[0].CreatorID}, nil)
	resp, err := s.usecase.GetTrendingResources(s.ctx, 20, nil)
	s.NoError(err)
	s.Equal(int64(1), resp.Total)
}

func (s *ResourceUsecaseTestSuite) TestGetTrendingResources_Error() {
	s.mockRepo.On("GetTrendingResources", mock.Anything, 20).Return(nil, errors.New("err"))
	_, err := s.usecase.GetTrendingResources(s.ctx, 20, nil)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestGetTopRatedResources_Success() {
	items := []resourcepkg.Resource{{ID: primitive.NewObjectID(), CreatorID: primitive.NewObjectID()}}
	s.mockRepo.On("GetTopRatedResources", mock.Anything, 20, "").Return(items, nil)
	s.mockUsers.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(userpkg.User{ID: items[0].CreatorID}, nil)
	resp, err := s.usecase.GetTopRatedResources(s.ctx, 20, "", nil)
	s.NoError(err)
	s.Equal(int64(1), resp.Total)
}

func (s *ResourceUsecaseTestSuite) TestGetTopRatedResources_Error() {
	s.mockRepo.On("GetTopRatedResources", mock.Anything, 20, "").Return(nil, errors.New("err"))
	_, err := s.usecase.GetTopRatedResources(s.ctx, 20, "", nil)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestGetUserBookmarkedResources_Success() {
	userID := primitive.NewObjectID()
	items := []resourcepkg.Resource{{ID: primitive.NewObjectID(), CreatorID: userID}}
	s.mockRepo.On("GetUserBookmarkedResources", mock.Anything, userID, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(items, int64(1), nil)
	s.mockUsers.On("FindByID", mock.Anything, userID.Hex()).Return(userpkg.User{ID: userID}, nil)
	s.mockRepo.On("IsResourceLikedByUser", mock.Anything, items[0].ID, userID).Return(false, nil)
	s.mockRepo.On("IsResourceBookmarkedByUser", mock.Anything, items[0].ID, userID).Return(true, nil)
	resp, err := s.usecase.GetUserBookmarkedResources(s.ctx, userID, resourcepkg.ResourcePagination{})
	s.NoError(err)
	s.Equal(int64(1), resp.Total)
}

func (s *ResourceUsecaseTestSuite) TestGetUserBookmarkedResources_Error() {
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetUserBookmarkedResources", mock.Anything, userID, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(nil, int64(0), errors.New("err"))
	_, err := s.usecase.GetUserBookmarkedResources(s.ctx, userID, resourcepkg.ResourcePagination{})
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestGetUserLikedResources_Success() {
	userID := primitive.NewObjectID()
	items := []resourcepkg.Resource{{ID: primitive.NewObjectID(), CreatorID: userID}}
	s.mockRepo.On("GetUserLikedResources", mock.Anything, userID, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(items, int64(1), nil)
	s.mockUsers.On("FindByID", mock.Anything, userID.Hex()).Return(userpkg.User{ID: userID}, nil)
	s.mockRepo.On("IsResourceLikedByUser", mock.Anything, items[0].ID, userID).Return(true, nil)
	s.mockRepo.On("IsResourceBookmarkedByUser", mock.Anything, items[0].ID, userID).Return(false, nil)
	resp, err := s.usecase.GetUserLikedResources(s.ctx, userID, resourcepkg.ResourcePagination{})
	s.NoError(err)
	s.Equal(int64(1), resp.Total)
}

func (s *ResourceUsecaseTestSuite) TestGetUserLikedResources_Error() {
	userID := primitive.NewObjectID()
	s.mockRepo.On("GetUserLikedResources", mock.Anything, userID, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(nil, int64(0), errors.New("err"))
	_, err := s.usecase.GetUserLikedResources(s.ctx, userID, resourcepkg.ResourcePagination{})
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestGetUpcomingOpportunities_Success() {
	items := []resourcepkg.Resource{{ID: primitive.NewObjectID(), CreatorID: primitive.NewObjectID()}}
	s.mockRepo.On("GetResourcesWithUpcomingDeadlines", mock.Anything, 7, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(items, int64(1), nil)
	s.mockUsers.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(userpkg.User{ID: items[0].CreatorID}, nil)
	resp, err := s.usecase.GetUpcomingOpportunities(s.ctx, 7, resourcepkg.ResourcePagination{}, nil)
	s.NoError(err)
	s.Equal(int64(1), resp.Total)
}

func (s *ResourceUsecaseTestSuite) TestGetUpcomingOpportunities_Error() {
	s.mockRepo.On("GetResourcesWithUpcomingDeadlines", mock.Anything, 7, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(nil, int64(0), errors.New("err"))
	_, err := s.usecase.GetUpcomingOpportunities(s.ctx, 7, resourcepkg.ResourcePagination{}, nil)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestGetResourcesWithDeadlines_Success() {
	items := []resourcepkg.Resource{{ID: primitive.NewObjectID(), CreatorID: primitive.NewObjectID()}}
	s.mockRepo.On("GetResources", mock.Anything, mock.AnythingOfType("resourcepkg.ResourceFilter"), mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(items, int64(1), nil)
	s.mockUsers.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(userpkg.User{ID: items[0].CreatorID}, nil)
	resp, err := s.usecase.GetResourcesWithDeadlines(s.ctx, resourcepkg.ResourcePagination{}, nil)
	s.NoError(err)
	s.Equal(int64(1), resp.Total)
}

func (s *ResourceUsecaseTestSuite) TestGetResourcesWithDeadlines_Error() {
	s.mockRepo.On("GetResources", mock.Anything, mock.AnythingOfType("resourcepkg.ResourceFilter"), mock.AnythingOfType("resourcepkg.ResourcePagination")).Return(nil, int64(0), errors.New("err"))
	_, err := s.usecase.GetResourcesWithDeadlines(s.ctx, resourcepkg.ResourcePagination{}, nil)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestReportResource_Success() {
	id := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("ReportResource", mock.Anything, id).Return(nil)
	err := s.usecase.ReportResource(s.ctx, id, primitive.NewObjectID(), "spam")
	s.NoError(err)
}

func (s *ResourceUsecaseTestSuite) TestReportResource_NotFound() {
	id := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(nil, errors.New("not found"))
	err := s.usecase.ReportResource(s.ctx, id, primitive.NewObjectID(), "spam")
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestVerifyResource_Success() {
	id := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id}, nil)
	s.mockRepo.On("VerifyResource", mock.Anything, id, mock.AnythingOfType("primitive.ObjectID")).Return(nil)
	err := s.usecase.VerifyResource(s.ctx, id, primitive.NewObjectID())
	s.NoError(err)
}

func (s *ResourceUsecaseTestSuite) TestVerifyResource_NotFound() {
	id := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(nil, errors.New("not found"))
	err := s.usecase.VerifyResource(s.ctx, id, primitive.NewObjectID())
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestSearchResources_EmptyQuery() {
	_, err := s.usecase.SearchResources(s.ctx, " ", resourcepkg.ResourceFilter{}, resourcepkg.ResourcePagination{}, nil)
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestGetResourceAnalytics_Unauthorized() {
	id := primitive.NewObjectID()
	creatorID := primitive.NewObjectID()
	actorID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id, CreatorID: creatorID}, nil)

	_, err := s.usecase.GetResourceAnalytics(s.ctx, id, actorID)
	s.Error(err)
	s.Contains(err.Error(), "unauthorized")
}

func (s *ResourceUsecaseTestSuite) TestGetResourceAnalytics_Success() {
	id := primitive.NewObjectID()
	creatorID := primitive.NewObjectID()
	s.mockRepo.On("GetResourceByID", mock.Anything, id).Return(&resourcepkg.Resource{ID: id, CreatorID: creatorID}, nil)
	stats := &resourcepkg.ResourceStats{ResourceID: id, ViewsCount: 10, LikesCount: 2, BookmarksCount: 3, SharesCount: 1, Rating: 4.5, RatingCount: 2}
	s.mockRepo.On("GetResourceStats", mock.Anything, id).Return(stats, nil)

	resp, err := s.usecase.GetResourceAnalytics(s.ctx, id, creatorID)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(10, resp.ViewsCount)
}

func (s *ResourceUsecaseTestSuite) TestGetUserResourceStats_Simple() {
	userID := primitive.NewObjectID()
	res1 := resourcepkg.Resource{ID: primitive.NewObjectID(), CreatorID: userID, ViewsCount: 10, LikesCount: 2, BookmarksCount: 1}
	res2 := resourcepkg.Resource{ID: primitive.NewObjectID(), CreatorID: userID, ViewsCount: 5, LikesCount: 1, BookmarksCount: 2}
	s.mockRepo.On("GetResourcesByCreator", mock.Anything, userID, mock.AnythingOfType("resourcepkg.ResourcePagination")).Return([]resourcepkg.Resource{res1, res2}, int64(2), nil)

	resp, err := s.usecase.GetUserResourceStats(s.ctx, userID)
	s.NoError(err)
	s.Equal(15, resp.TotalViews)
	s.Equal(3, resp.TotalLikes)
	s.Equal(3, resp.TotalBookmarks)
}

func (s *ResourceUsecaseTestSuite) TestValidateResourceType_Invalid() {
	err := s.usecase.ValidateResourceType("invalid-type")
	s.Error(err)
}

func (s *ResourceUsecaseTestSuite) TestCreateResource_FailsWhenUserLookupFails() {
	creatorID := primitive.NewObjectID()
	s.mockUsers.On("FindByID", mock.Anything, creatorID.Hex()).Return(userpkg.User{}, errors.New("db down"))

	_, err := s.usecase.CreateResource(s.ctx, resourcepkg.CreateResourceRequest{Title: "T", Description: "D", Content: "C", Type: "guide", Category: "Academic Success"}, creatorID)
	s.Error(err)
	s.Contains(err.Error(), "failed to get creator")
}
