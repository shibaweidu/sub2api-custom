package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScheduledTestFailuresAutomaticallyUnpublished(t *testing.T) {
	for _, status := range []string{"failed", "error", "success", "passed"} {
		t.Run(status, func(t *testing.T) {
			for _, publish := range []bool{false, true} {
				repo := &runnerResultRepoStub{}
				svc := NewScheduledTestService(nil, repo)
				result := &ScheduledTestResult{Status: status, Published: publish}
				require.NoError(t, svc.SaveResult(context.Background(), 7, 50, result))
				expected := publish && (status == "success" || status == "passed")
				require.Equal(t, expected, repo.created[0].Published)
				// Also cover completing a row that was published while running.
				result.ID, result.Published = repo.created[0].ID, publish
				require.NoError(t, svc.CompleteResult(context.Background(), 50, result))
				require.Equal(t, expected, repo.created[0].Published)
				require.Equal(t, status, repo.created[0].Status, "admin keeps the failure for diagnosis")
			}
		})
	}
}
