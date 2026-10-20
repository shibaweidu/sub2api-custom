//go:build unit

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestScheduledTestPlanRepositoryCreatePersistsName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &scheduledTestPlanRepository{db: db}
	groupID := int64(3)
	createdAt := time.Now()
	nextRun := createdAt.Add(time.Minute)
	mock.ExpectQuery(`SELECT COUNT\(\*\),COUNT\(DISTINCT platform\) FROM groups`).
		WithArgs("{3}").
		WillReturnRows(sqlmock.NewRows([]string{"count", "platforms"}).AddRow(1, 1))
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)INSERT INTO scheduled_test_plans \(name, sort_order, account_id`).
		WithArgs("nightly candy", 0, nil, groupID, nil, "candy", "all_accounts", "model", "", "*/5 * * * *", true, 20, false, nextRun, "{2}", sqlmock.AnyArg(), "{3}", "", sqlmock.AnyArg(), sqlmock.AnyArg(), true, "{7}").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "sort_order", "account_id", "group_id", "test_definition_id", "test_type", "target_mode", "model_id", "reasoning_effort", "cron_expression", "enabled", "max_results", "auto_recover", "last_run_at", "next_run_at", "created_at", "updated_at", "test_definition_ids", "protection", "group_ids", "migration_note", "schedule_config", "account_ids", "auto_publish_results", "excluded_account_ids",
		}).AddRow(10, "nightly candy", 0, nil, groupID, nil, "candy", "all_accounts", "model", "", "*/5 * * * *", true, 20, false, nil, nextRun, createdAt, createdAt, "{2}", "{}", "{3}", "", nil, "{}", true, "{7}"))
	mock.ExpectCommit()

	got, err := repo.Create(context.Background(), &service.ScheduledTestPlan{
		Name: "nightly candy", ExcludedAccountIDs: []int64{7}, GroupID: &groupID, GroupIDs: []int64{groupID}, TestDefinitionIDs: []int64{2}, TargetMode: "all_accounts", TestType: "candy", ModelID: "model", CronExpression: "*/5 * * * *", Enabled: true, MaxResults: 20, NextRunAt: &nextRun, AutoPublishResults: true,
	})
	require.NoError(t, err)
	require.Equal(t, "nightly candy", got.Name)
	require.Equal(t, []int64{3}, got.GroupIDs)
	require.Equal(t, []int64{7}, got.ExcludedAccountIDs)
	require.Equal(t, []int64{2}, got.TestDefinitionIDs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduledTestResultRepositoryListIncludesDisplayNames(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &scheduledTestResultRepository{db: db}
	createdAt := time.Now()
	mock.ExpectQuery(`(?s)SELECT r\.id, r\.plan_id.*COALESCE\(a.name, ''\).*FROM scheduled_test_results r.*LEFT JOIN accounts a ON a.id = r.account_id.*ORDER BY r\.started_at DESC, r\.id DESC`).
		WithArgs(int64(10), 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "plan_name", "test_name", "test_order", "group_name", "plan_order", "target_mode", "status", "response_text", "output_kind", "output_html", "output_numeric", "account_id", "model_id", "reasoning_effort", "group_id", "error_message", "latency_ms", "started_at", "finished_at", "created_at", "test_definition_id", "account_name", "published",
		}).AddRow(1, 10, "nightly candy", "糖果数字测试", 0, "公开组", 12, "all_accounts", "success", "答案：29", "number", "", 29.0, 7, "model", "high", 3, "", 42, createdAt, createdAt, createdAt, 2, "Account Seven", true))
	mock.ExpectQuery(`(?s)SELECT id, protection_decision.*FROM scheduled_test_results`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "protection_decision", "execution_info"}))

	results, err := repo.ListByPlanID(context.Background(), 10, 20)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "nightly candy", results[0].PlanName)
	require.Equal(t, "糖果数字测试", results[0].TestName)
	require.Equal(t, "公开组", results[0].GroupName)
	require.Equal(t, 12, results[0].PlanOrder)
	require.Equal(t, "high", results[0].ReasoningEffort)
	require.Equal(t, "Account Seven", results[0].AccountName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduledTestResultRepositoryPersistsReasoningEffort(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &scheduledTestResultRepository{db: db}
	now := time.Now()
	accountID, groupID := int64(7), int64(3)
	input := &service.ScheduledTestResult{
		PlanID: 10, Status: "running", OutputKind: "text", AccountID: &accountID,
		GroupID: &groupID, ModelID: "gpt-6-astra", ReasoningEffort: "ultra", StartedAt: now, FinishedAt: now,
	}
	mock.ExpectQuery(`(?s)INSERT INTO scheduled_test_results .*reasoning_effort.*RETURNING .*reasoning_effort`).
		WithArgs(int64(10), "running", "", "text", "", nil, accountID, "gpt-6-astra", "ultra", groupID, "", int64(0), now, now, nil, "", "", []byte("null"), false).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "status", "response_text", "output_kind", "output_html", "output_numeric", "account_id", "model_id", "reasoning_effort", "group_id", "error_message", "latency_ms", "started_at", "finished_at", "created_at", "test_definition_id", "target_mode", "published",
		}).AddRow(1, 10, "running", "", "text", "", nil, accountID, "gpt-6-astra", "ultra", groupID, "", 0, now, now, now, nil, "", false))
	created, err := repo.Create(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, "ultra", created.ReasoningEffort)

	created.Status = "success"
	created.ResponseText = "done"
	mock.ExpectExec(`(?s)UPDATE scheduled_test_results .*reasoning_effort = \$9`).
		WithArgs(int64(1), "success", "done", "text", "", nil, accountID, "gpt-6-astra", "ultra", groupID, "", int64(0), now, now, []byte("null"), false).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), created))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduledTestPlanAccountSelectionFiltersTargets(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := &scheduledTestResultRepository{db: db}
	groupIDs, accountIDs := []int64{3}, []int64{31, 32}
	mock.ExpectQuery(`(?s)SELECT a\.id FROM accounts a.*COALESCE\(cardinality\(\$4::bigint\[\]\),0\)=0 OR a\.id=ANY\(\$4::bigint\[\]\)`).
		WithArgs(nil, "{3}", false, "{31,32}", "{31}").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(32))
	ids, err := repo.listPlanAccountIDs(context.Background(), &service.ScheduledTestPlan{GroupIDs: groupIDs, AccountIDs: accountIDs, ExcludedAccountIDs: []int64{31}}, nil, false)
	require.NoError(t, err)
	require.Equal(t, []int64{32}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduledTestResultRepositorySetPublished(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := &scheduledTestResultRepository{db: db}
	mock.ExpectExec(`UPDATE scheduled_test_results SET published=\$2 WHERE id=\$1 AND \(NOT \$2 OR status IN \('success', 'passed'\)\)`).
		WithArgs(int64(7), true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.SetPublished(context.Background(), 7, true))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduledTestResultRepositoryGroupViewsIncludePublishedRunHistory(t *testing.T) {
	columns := []string{"id", "plan_id", "plan_name", "test_name", "test_order", "group_name", "plan_order", "target_mode", "status", "response_text", "output_kind", "output_html", "output_numeric", "account_id", "model_id", "reasoning_effort", "group_id", "error_message", "latency_ms", "started_at", "finished_at", "created_at", "test_definition_id", "group_order", "account_schedulable", "test_key", "quality_verdict"}
	newRows := func() *sqlmock.Rows { return sqlmock.NewRows(columns) }
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := &scheduledTestResultRepository{db: db}

	mock.ExpectQuery(`(?s)WITH projected_results AS .*FROM visible_results vr\s+WHERE vr\.group_id = \$2 AND vr.status IN \('success', 'passed'\)\s+ORDER BY vr\.started_at DESC, vr\.id DESC\s+LIMIT \$3`).
		WithArgs(int64(5), int64(8), 4).
		WillReturnRows(newRows())
	recent, err := repo.ListVisibleRecentForGroup(context.Background(), 5, 8, 4)
	require.NoError(t, err)
	require.Empty(t, recent)

	mock.ExpectQuery(`(?s)WITH projected_results AS .*PARTITION BY vr\.group_id, vr\.result_target_key, vr\.test_definition_id, vr\.model_id, vr\.reasoning_effort.*WHERE vr\.group_id = \$2.*WHERE history_rank = 1`).
		WithArgs(int64(5), int64(8)).
		WillReturnRows(newRows())
	latest, err := repo.ListVisibleForGroup(context.Background(), 5, 8)
	require.NoError(t, err)
	require.Empty(t, latest)

	historyQuery := `(?s)WITH projected_results AS .*history_anchor AS \(\s+SELECT \* FROM visible_results\s+WHERE id = \$2\s+\), history_series AS .*WHERE vr\.status IN \('success', 'passed', 'failed', 'error', 'pending', 'running'\).*`
	mock.ExpectQuery(historyQuery+`SELECT EXISTS`).
		WithArgs(int64(5), int64(91), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"anchor_visible", "cursor_valid"}).AddRow(true, true))
	mock.ExpectQuery(historyQuery+`SELECT vr\.id,`).
		WithArgs(int64(5), int64(91), int64(0), 20).
		WillReturnRows(newRows().AddRow(91, 3, "nightly candy", "image check", 0, "Pro", 0, "all_accounts", "running", "", "html", "", nil, 17, "model", "high", 8, "", 0, time.Now(), time.Time{}, time.Now(), 2, 0, true, "pelican", ""))
	history, err := repo.ListVisibleHistory(context.Background(), 5, 91, 0, 20)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, "running", history[0].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduledTestResultRepositoryVisibleUsesResultReasoningEffort(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &scheduledTestResultRepository{db: db}
	now := time.Now()
	mock.ExpectQuery(`(?s)WITH projected_results AS .*r.published AND r.status IN \('success', 'passed', 'pending', 'running'\).*ranked_results AS.*PARTITION BY vr.group_id, vr.result_target_key, vr.test_definition_id, vr.model_id, vr.reasoning_effort.*ORDER BY \(vr.status IN \('success', 'passed'\)\) DESC.*WHERE history_rank = 1.*ORDER BY vr.account_schedulable DESC NULLS LAST.*vr.started_at DESC, vr.id DESC`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "plan_name", "test_name", "test_order", "group_name", "plan_order", "target_mode", "status", "response_text", "output_kind", "output_html", "output_numeric", "account_id", "model_id", "reasoning_effort", "group_id", "error_message", "latency_ms", "started_at", "finished_at", "created_at", "test_definition_id", "group_order", "account_schedulable", "test_key", "quality_verdict",
		}).AddRow(1, 10, "nightly candy", "糖果数字测试", 0, "公开组", 12, "all_accounts", "success", "29", "number", "", 29.0, 7, "model", "medium", 3, "", 42, now, now, now, 2, 0, true, "candy", "fail"))
	results, err := repo.ListVisible(context.Background(), 5, 20)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "medium", results[0].ReasoningEffort)
	require.Equal(t, "candy", results[0].TestKey)
	require.Equal(t, "fail", results[0].QualityVerdict)
	require.NotNil(t, results[0].GroupOrder)
	require.Zero(t, *results[0].GroupOrder)
	require.NotNil(t, results[0].AccountSchedulable)
	require.True(t, *results[0].AccountSchedulable)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduledTestResultRepositoryDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &scheduledTestResultRepository{db: db}
	mock.ExpectQuery(`SELECT plan_id,account_id FROM scheduled_test_results WHERE id=\$1`).
		WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"plan_id", "account_id"}).AddRow(8, nil))
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT enabled, protection FROM scheduled_test_plans WHERE id=\$1 FOR NO KEY UPDATE`).
		WithArgs(int64(8)).WillReturnRows(sqlmock.NewRows([]string{"enabled", "protection"}).AddRow(true, "{}"))
	mock.ExpectExec(`DELETE FROM scheduled_test_results WHERE id = \$1`).
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	require.NoError(t, repo.Delete(context.Background(), 42))
	require.NoError(t, mock.ExpectationsWereMet())
}
