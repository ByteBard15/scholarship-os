package workflow

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// This opt-in test creates an isolated schema on a disposable LOCAL PostgreSQL
// instance and applies the real migrations, including all CHECK constraints.
func TestProposalMaterializationWithResearchDeadlineValues(t *testing.T) {
	dsn := os.Getenv("WORKFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set WORKFLOW_TEST_DATABASE_URL to a disposable local PostgreSQL instance")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || (parsed.Hostname() != "127.0.0.1" && parsed.Hostname() != "localhost") {
		t.Fatal("integration test requires a local disposable database")
	}
	base, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open disposable database:", err)
	}
	baseSQL, _ := base.DB()
	defer baseSQL.Close()
	schema := "proposal_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := base.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	defer base.Exec("DROP SCHEMA " + schema + " CASCADE")
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: parsed.String(), PreferSimpleProtocol: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	migrations, err := filepath.Glob("../../../../../db/migrations/*.up.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatal("could not find migrations", err)
	}
	for _, path := range migrations {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(body)).Error; err != nil {
			t.Fatalf("migration %s: %v", filepath.Base(path), err)
		}
	}
	owner := user.User{Email: uuid.NewString() + "@example.test", PasswordHash: "test-fixture"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	parent := profile.ApplicantProfile{UserID: owner.ID, Name: "Master", ProfileType: profile.ProfileTypeMaster}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatal(err)
	}
	task := ResearchTask{UserID: owner.ID, ProfileID: &parent.ID, Title: "Biomedical engineering", TaskType: "scholarship_research", Status: "review_required"}
	if err := db.Create(&task).Error; err != nil {
		t.Fatal(err)
	}
	run := application.ResearchRun{ResearchTaskID: &task.ID, Status: "review_required", Trigger: "agent", ResearchType: "scholarship_research"}
	if err := db.Create(&run).Error; err != nil {
		t.Fatal(err)
	}
	source := application.ResearchSource{ResearchRunID: run.ID, URL: "https://example.edu/deadline", SourceType: "admission", RetrievedAt: time.Now().UTC()}
	if err := db.Create(&source).Error; err != nil {
		t.Fatal(err)
	}
	proposal := ApplicationProposal{UserID: owner.ID, ResearchTaskID: task.ID, ResearchRunID: &run.ID, ParentProfileID: &parent.ID, Name: "BME MSc", Status: "approved", Priority: "highest", ProposedInstitution: profile.JSON(`{"name":"Example University","country":"Canada"}`), ProposedProgramme: profile.JSON(`{"name":"Biomedical Engineering","degreeLevel":"master"}`)}
	if err := db.Create(&proposal).Error; err != nil {
		t.Fatal(err)
	}
	findings := []application.ResearchFinding{
		{ResearchRunID: run.ID, SourceID: &source.ID, Category: "deadline", Field: "internationalAdmissionDeadline", Value: profile.JSON(`{"rawText":"28 February 2027, only for eligible applicants","datePrecision":"date","localDate":"2027-02-28"}`), VerificationStatus: "verified", ReviewStatus: "pending"},
		{ResearchRunID: run.ID, SourceID: &source.ID, Category: "funding", Field: "fundingScope", Value: profile.JSON(`{"currency":"CAD","coverage":"Fees and living included; confirm net amount"}`), VerificationStatus: "supported", ReviewStatus: "pending"},
		{ResearchRunID: run.ID, SourceID: &source.ID, Category: "requirement", Field: "academicEntry", Value: profile.JSON(`{"entry":"University must assess degree equivalence"}`), VerificationStatus: "supported", ReviewStatus: "pending"},
		{ResearchRunID: run.ID, SourceID: &source.ID, Category: "requirement", Field: "withdrawnRequirement", Value: profile.JSON(`{"category":"academic","title":"Rejected evidence"}`), VerificationStatus: "verified", ReviewStatus: "rejected"},
	}
	if err := db.Create(&findings).Error; err != nil {
		t.Fatal(err)
	}
	// Reproduce the former insert against the actual precision CHECK constraint.
	oldInsert := db.Transaction(func(tx *gorm.DB) error {
		temp := application.Application{UserID: owner.ID, ApplicantProfileID: parent.ID, Name: "Old insertion", Status: application.StatusResearchComplete, ResearchStatus: application.ResearchComplete}
		if err := tx.Create(&temp).Error; err != nil {
			return err
		}
		return tx.Create(&application.ApplicationDeadline{ApplicationID: temp.ID, DeadlineType: "", Title: "", DatePrecision: "date"}).Error
	})
	var pgErr *pgconn.PgError
	if !errors.As(oldInsert, &pgErr) || pgErr.ConstraintName != "application_deadlines_precision_check" {
		t.Fatalf("expected precision failure; got %v", oldInsert)
	}
	repo := NewGORMRepository(db)
	created, err := repo.CreateApplicationFromProposal(context.Background(), &proposal, &task, &parent)
	if err != nil {
		t.Fatalf("materialization failed: %v", err)
	}
	var deadlines []application.ApplicationDeadline
	if err := db.Where("application_id = ?", created.ID).Find(&deadlines).Error; err != nil {
		t.Fatal(err)
	}
	if len(deadlines) != 1 || deadlines[0].DatePrecision != "day" || deadlines[0].DeadlineAt != nil || deadlines[0].SourceID == nil || *deadlines[0].SourceID != source.ID || deadlines[0].RawDeadlineText == nil {
		t.Fatalf("deadline lost precision/provenance: %+v", deadlines)
	}
	for i := range findings {
		var got application.ResearchFinding
		if err := db.First(&got, "id = ?", findings[i].ID).Error; err != nil {
			t.Fatal(err)
		}
		want := "pending"
		if i == 0 {
			want = "accepted"
		}
		if i == 3 {
			want = "rejected"
		}
		if got.ReviewStatus != want || string(got.Value) == "" {
			t.Fatalf("finding %d review status: %s", i, got.ReviewStatus)
		}
	}
	var fundingCount, requirementCount int64
	db.Model(&application.ApplicationFunding{}).Where("application_id = ?", created.ID).Count(&fundingCount)
	db.Model(&application.ApplicationRequirement{}).Where("application_id = ?", created.ID).Count(&requirementCount)
	if fundingCount != 0 || requirementCount != 0 {
		t.Fatal("created blank records from unsupported evidence")
	}
	if _, err := repo.CreateApplicationFromProposal(context.Background(), &proposal, &task, &parent); err != ErrInvalidProposalState {
		t.Fatalf("duplicate materialization allowed: %v", err)
	}
}
