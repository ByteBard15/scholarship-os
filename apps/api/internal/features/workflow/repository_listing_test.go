package workflow

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestResearchListingSQLFiltersOrdersAndPaginates(t *testing.T) {
	var output bytes.Buffer
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true,
		Logger: logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info}),
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := &GORMRepository{db: db}
	owner, parent := uuid.New(), uuid.New()
	priority := "normal"
	filters := ResearchTaskFilters{UserID: &owner, ParentTaskID: &parent, Priority: &priority, SortBy: "priority", OldestFirst: true, Limit: 10, Offset: 20}
	if _, err := repo.ListResearchTasks(context.Background(), filters); err != nil {
		t.Fatal(err)
	}
	sql := output.String()
	for _, want := range []string{owner.String(), parent.String(), "COALESCE(priority, 'normal') = 'normal'", "WHEN 'urgent' THEN 1", "due_at ASC NULLS LAST", "created_at ASC", "id ASC", "LIMIT 10 OFFSET 20"} {
		if !strings.Contains(sql, want) {
			t.Fatalf("listing missing %q: %s", want, sql)
		}
	}
	output.Reset()
	if _, err := repo.CountResearchTasks(context.Background(), filters); err != nil {
		t.Fatal(err)
	}
	sql = output.String()
	if !strings.Contains(sql, owner.String()) || !strings.Contains(sql, parent.String()) || strings.Contains(sql, "LIMIT") || strings.Contains(sql, "OFFSET") {
		t.Fatalf("filtered total must ignore pagination: %s", sql)
	}
	output.Reset()
	filters.SortBy = "dueAt"
	if _, err := repo.ListResearchTasks(context.Background(), filters); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "ORDER BY due_at ASC NULLS LAST") {
		t.Fatal(output.String())
	}
}

func TestDeadlineQueryUsesLatestRunAndSourcedNonRejectedFindings(t *testing.T) {
	var output bytes.Buffer
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, Logger: logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info})})
	if err != nil {
		t.Fatal(err)
	}
	repo := &GORMRepository{db: db}
	id := uuid.New()
	if _, err := repo.ListTaskDeadlineFindings(context.Background(), []uuid.UUID{id}); err != nil {
		t.Fatal(err)
	}
	sql := output.String()
	for _, want := range []string{"DISTINCT ON (research_task_id)", id.String(), "created_at DESC, id DESC", "JOIN research_sources", "sources.research_run_id = runs.id", "runs.status IN ('review_required','completed')", "findings.review_status <> 'rejected'"} {
		if !strings.Contains(sql, want) {
			t.Fatalf("deadline query missing %q: %s", want, sql)
		}
	}
}

func TestResearchProposalFiltersUseExistsForRowsAndTotals(t *testing.T) {
	for _, state := range []string{"any", "none", "approved"} {
		t.Run(state, func(t *testing.T) {
			var output bytes.Buffer
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, Logger: logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info})})
			if err != nil {
				t.Fatal(err)
			}
			repo := &GORMRepository{db: db}
			owner := uuid.New()
			filters := ResearchTaskFilters{UserID: &owner, ProposalState: state, Limit: 20}
			for _, count := range []bool{false, true} {
				output.Reset()
				if count {
					_, err = repo.CountResearchTasks(context.Background(), filters)
				} else {
					_, err = repo.ListResearchTasks(context.Background(), filters)
				}
				if err != nil {
					t.Fatal(err)
				}
				sql := output.String()
				for _, want := range []string{owner.String(), "EXISTS (SELECT 1", "application_proposals.research_task_id = research_tasks.id", "application_proposals.user_id = research_tasks.user_id"} {
					if !strings.Contains(sql, want) {
						t.Fatalf("missing %q: %s", want, sql)
					}
				}
				if strings.Contains(sql, "NOT EXISTS") != (state == "none") {
					t.Fatal(sql)
				}
				if strings.Contains(sql, "application_proposals.status = 'approved'") != (state == "approved") {
					t.Fatal(sql)
				}
				if strings.Contains(sql, "JOIN") || strings.Contains(sql, "application_id IS NULL") {
					t.Fatalf("must not duplicate tasks or hide materialized approvals: %s", sql)
				}
			}
		})
	}
}
