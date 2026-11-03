package writing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Run against a disposable database with migrations 1–17 applied.
func TestWritingRepositoryPreparationPostgres(t *testing.T) {
	dsn := os.Getenv("AGENT_ACCESS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("AGENT_ACCESS_TEST_DATABASE_URL not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	owner, pid, snapshot := uuid.New(), uuid.New(), uuid.New()
	for _, query := range []struct {
		sql  string
		args []any
	}{
		{"INSERT INTO users(id,email,password_hash) VALUES (?,?,?)", []any{owner, owner.String() + "@example.test", "test-only"}},
		{"INSERT INTO applicant_profiles(id,user_id,name,profile_type) VALUES (?,?,?,?)", []any{pid, owner, "Master", "master"}},
	} {
		if err := tx.Exec(query.sql, query.args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Create(&profile.ProfileSnapshot{ID: snapshot, ProfileID: pid, Version: 1, Snapshot: profile.JSON(`{}`)}).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewGORMRepository(tx)
	ctx := context.Background()
	sample := &WritingSample{ID: uuid.New(), UserID: owner, Title: "Base Recommendation Letter #2", DocumentType: TypeEmail, Content: "Sensitive full text", Origin: OriginAgent, Status: StatusSuggested, SourceProfileSnapshotID: &snapshot, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.CreateSample(ctx, sample, nil, nil); err != nil {
		t.Fatal(err)
	}
	metadata, err := repo.ListSampleMetadata(ctx, owner, SampleMetadataQuery{Title: "base recommendation letter #2", Limit: 10})
	if err != nil || len(metadata) != 1 || metadata[0].ID != sample.ID {
		t.Fatalf("metadata: %+v %v", metadata, err)
	}
	if rows, err := repo.ListSampleMetadata(ctx, uuid.New(), SampleMetadataQuery{Limit: 10}); err != nil || len(rows) != 0 {
		t.Fatal("metadata leaked", err)
	}
	// Simulate a human approving after the agent read the suggested record.
	if err := tx.Model(&WritingSample{}).Where("id = ?", sample.ID).Update("status", StatusApproved).Error; err != nil {
		t.Fatal(err)
	}
	sample.Content = "Stale overwrite"
	if err := repo.UpdateSuggestedSample(ctx, sample, nil); err != ErrDraftReadOnly {
		t.Fatal("concurrent approval overwritten", err)
	}
	stored, err := repo.GetSample(ctx, sample.ID)
	if err != nil || stored.Status != StatusApproved || stored.Content != "Sensitive full text" || stored.SourceProfileSnapshotID == nil {
		t.Fatalf("stored draft: %+v %v", stored, err)
	}
}
