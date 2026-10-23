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
