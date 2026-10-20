package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type userTestGroupResultsRepo struct {
	service.ScheduledTestResultRepository
	allRows    []*service.ScheduledTestResult
	latestRows []*service.ScheduledTestResult
	allCalls   int
	recentCall struct {
		groupID int64
		limit   int
	}
}

func (r *userTestGroupResultsRepo) ListVisibleForGroup(_ context.Context, _, groupID int64) ([]*service.ScheduledTestResult, error) {
	r.allCalls++
	return r.allRows, nil
}

func (r *userTestGroupResultsRepo) ListVisibleRecentForGroup(_ context.Context, _, groupID int64, limit int) ([]*service.ScheduledTestResult, error) {
	r.recentCall.groupID, r.recentCall.limit = groupID, limit
	return r.latestRows, nil
}

func TestListTestResultsForGroupReturnsAllRowsWithOpaqueSetKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountA, accountB := int64(321), int64(654)
	repo := &userTestGroupResultsRepo{allRows: []*service.ScheduledTestResult{
		{ID: 10, AccountID: &accountA, TargetMode: "all_accounts", GroupID: ptrInt64(8), Status: "success", OutputKind: "html", StartedAt: time.Now()},
		{ID: 11, AccountID: &accountA, TargetMode: "all_accounts", GroupID: ptrInt64(8), Status: "success", OutputKind: "html", StartedAt: time.Now()},
		{ID: 12, AccountID: &accountB, TargetMode: "all_accounts", GroupID: ptrInt64(8), Status: "failed", OutputKind: "html", StartedAt: time.Now()},
	}}
	h := &UserHandler{scheduledTestSvc: service.NewScheduledTestService(nil, repo)}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42}) })
	router.GET("/test-results", h.ListTestResults)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/test-results?group_id=8", nil))

	require.Equal(t, http.StatusOK, response.Code)
	var rows []userQualityResultSummary
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &rows))
	require.Len(t, rows, 3)
	require.Equal(t, int64(10), *rows[0].ResultSetID)
	require.Equal(t, int64(10), *rows[1].ResultSetID)
	require.Equal(t, int64(12), *rows[2].ResultSetID)
	require.Equal(t, "all_accounts", rows[0].TargetMode)
	require.NotContains(t, response.Body.String(), "account_id")
	require.NotContains(t, response.Body.String(), "account_name")
	require.Equal(t, 1, repo.allCalls)
	require.Equal(t, int64(0), repo.recentCall.groupID)
}

func TestListTestResultsForGroupLimitUsesCompactLatestQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &userTestGroupResultsRepo{}
	h := &UserHandler{scheduledTestSvc: service.NewScheduledTestService(nil, repo)}
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
	c.Request = httptest.NewRequest(http.MethodGet, "/test-results?group_id=8&limit=4", nil)
	h.ListTestResults(c)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, 0, repo.allCalls)
	require.Equal(t, int64(8), repo.recentCall.groupID)
	require.Equal(t, 4, repo.recentCall.limit)
}

func ptrInt64(value int64) *int64 { return &value }
