//go:build integration

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func testScheduledTestMonitorGroups(t *testing.T, ctx context.Context, db *sql.DB, plans service.ScheduledTestPlanRepository, results *scheduledTestResultRepository, candyID int64) {
	_, err := db.ExecContext(ctx, `INSERT INTO groups(id,name,is_exclusive) VALUES
(99011,'Monitor public',false),(99012,'Monitor private',true),(99013,'Monitor allowed',true);
INSERT INTO accounts(id,name,schedulable) VALUES (99001,'Private account name',false);
INSERT INTO account_groups(account_id,group_id) VALUES (99001,99011),(99001,99012),(99001,99013);
INSERT INTO user_allowed_groups(user_id,group_id) VALUES (100,99013)`)
	require.NoError(t, err)
	groupID, accountID := int64(99011), int64(99001)
	plan, err := plans.Create(ctx, &service.ScheduledTestPlan{
		Name: "Monitor membership", GroupIDs: []int64{groupID}, GroupID: &groupID,
		TestDefinitionID: &candyID, TestDefinitionIDs: []int64{candyID},
		TargetMode: "all_accounts", ModelID: "model", CronExpression: "0 * * * *", MaxResults: 10,
	})
	require.NoError(t, err)
	row, err := results.Create(ctx, &service.ScheduledTestResult{
		PlanID: plan.ID, TestDefinitionID: &candyID, AccountID: &accountID, GroupID: &groupID,
		TargetMode: "all_accounts", ModelID: "model", OutputKind: "number", ResponseText: "29", Status: "success",
		Published: true,
		StartedAt: time.Now(), FinishedAt: time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, results.RecordDecision(ctx, row, service.ScheduledTestProtectionDecision{
		Verdict: "fail", Reason: "private diagnostic", AddedGroupIDs: []int64{99012},
	}))
	svc := service.NewScheduledTestService(plans, results)
	assertGroups := func(want []int64) {
		t.Helper()
		visible, err := svc.ListVisibleGroupResults(ctx, 100)
		require.NoError(t, err)
		groups := []int64{}
		for _, got := range visible {
			if got.ID != row.ID {
				continue
			}
			groups = append(groups, *got.GroupID)
			require.Equal(t, "candy", got.TestKey)
			require.Equal(t, "fail", got.QualityVerdict, "request success is not a quality pass")
			require.NotNil(t, got.AccountSchedulable)
			require.False(t, *got.AccountSchedulable)
			encoded, err := json.Marshal(got)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "private diagnostic")
			require.NotContains(t, string(encoded), "Private account name")
			require.NotContains(t, string(encoded), "added_group_ids")
		}
		require.ElementsMatch(t, want, groups)
	}
	assertGroups([]int64{99011, 99013})
	ordinary, err := results.ListVisible(ctx, 100, 3)
	require.NoError(t, err)
	count := 0
	for _, got := range ordinary {
		if got.ID == row.ID {
			count++
			require.Equal(t, groupID, *got.GroupID)
		}
	}
	require.Equal(t, 1, count, "ordinary results keep their original grouping")
	t.Run("linked group shares recent work and history with publication controls", func(t *testing.T) {
		at := time.Now().UTC().Truncate(time.Microsecond)
		var latestID int64
		for i := 0; i < 4; i++ {
			copy := *row
			copy.StartedAt, copy.FinishedAt = at.Add(time.Duration(i)*time.Second), at.Add(time.Duration(i)*time.Second)
			created, err := results.Create(ctx, &copy)
			require.NoError(t, err)
			latestID = created.ID
		}
		recent, err := results.ListVisibleRecentForGroup(ctx, 100, 99013, 4)
		require.NoError(t, err)
		require.Len(t, recent, 4, "a secondary group membership can fill all four previews")
		require.Equal(t, latestID, recent[0].ID)
		for _, got := range recent {
			require.Equal(t, int64(99013), *got.GroupID)
		}
		history, err := results.ListVisibleHistoryForGroup(ctx, 100, 99013, latestID, 0, 20)
		require.NoError(t, err)
		require.Len(t, history, 5, "the linked page reads the same published history")
		require.NoError(t, results.SetPublished(ctx, latestID, false))
		recent, err = results.ListVisibleRecentForGroup(ctx, 100, 99013, 4)
		require.NoError(t, err)
		require.Len(t, recent, 4)
		require.NotEqual(t, latestID, recent[0].ID, "hidden work does not displace published previews")
		_, err = results.ListVisibleHistoryForGroup(ctx, 100, 99013, latestID, 0, 20)
		require.ErrorIs(t, err, sql.ErrNoRows, "a direct URL cannot reveal unpublished work")
		require.NoError(t, results.SetPublished(ctx, latestID, true))
		private, err := results.ListVisibleRecentForGroup(ctx, 100, 99012, 4)
		require.NoError(t, err)
		require.Empty(t, private, "secondary memberships still require group access")
	})
	_, err = db.ExecContext(ctx, `DELETE FROM account_groups WHERE account_id=99001 AND group_id=99011;
DELETE FROM user_allowed_groups WHERE user_id=100 AND group_id=99013`)
	require.NoError(t, err)
	assertGroups(nil)
}
