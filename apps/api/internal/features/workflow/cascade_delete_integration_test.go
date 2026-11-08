package workflow

import (
	"context"
	"github.com/byte/scholarship-os/apps/api/internal/features/application"
	"github.com/byte/scholarship-os/apps/api/internal/features/profile"
	"github.com/byte/scholarship-os/apps/api/internal/features/user"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResearchCascadeDeletion(t *testing.T) {
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

	owner := user.User{Email: uuid.NewString() + "@example.test", PasswordHash: "test"}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(db.Create(&owner).Error)
	parent := profile.ApplicantProfile{UserID: owner.ID, Name: "Master", ProfileType: profile.ProfileTypeMaster}
	must(db.Create(&parent).Error)
	tailored := profile.ApplicantProfile{UserID: owner.ID, Name: "Application", ProfileType: profile.ProfileTypeApplication, ParentProfileID: &parent.ID}
	must(db.Create(&tailored).Error)
	app := application.Application{UserID: owner.ID, ApplicantProfileID: tailored.ID, Name: "Cascade application", Status: application.StatusResearchComplete, ResearchStatus: application.ResearchComplete}
	must(db.Create(&app).Error)
	task := ResearchTask{UserID: owner.ID, Title: "Root", TaskType: "scholarship_research", Status: "draft", TargetApplicationID: &app.ID}
	must(db.Create(&task).Error)
	child := ResearchTask{UserID: owner.ID, Title: "Child", TaskType: "scholarship_research", Status: "running", ParentTaskID: &task.ID}
	must(db.Create(&child).Error)
	repo := NewGORMRepository(db)
	ctx := context.Background()
	preview, err := repo.PreviewResearchDeletion(ctx, &task)
	must(err)
	if len(preview.Blockers) == 0 {
		t.Fatal("running child must block deletion")
	}
	if err := repo.CascadeDeleteResearch(ctx, &task, preview.ConfirmationToken); err == nil {
		t.Fatal("running deletion succeeded")
	}
	must(db.Model(&child).Update("status", "draft").Error)
	must(db.Delete(&child).Error)
	preview, err = repo.PreviewResearchDeletion(ctx, &task)
	must(err)
	if len(preview.Tasks) != 2 || len(preview.Applications) != 0 {
		t.Fatal("must include hidden child without deleting target application")
	}
	oldToken := preview.ConfirmationToken
	proposal := ApplicationProposal{UserID: owner.ID, ResearchTaskID: child.ID, ApplicationID: &app.ID, Name: "Proposal", Status: "approved", Priority: "high"}
	must(db.Create(&proposal).Error)
	if err := repo.CascadeDeleteResearch(ctx, &task, oldToken); err == nil {
		t.Fatal("stale token accepted")
	}
	run := application.ResearchRun{ResearchTaskID: &child.ID, Status: "review_required", Trigger: "agent", ResearchType: "scholarship_research"}
	must(db.Create(&run).Error)

	writingID := uuid.New()
	must(db.Exec(`INSERT INTO writing_samples (id,user_id,application_id,title,document_type,content,created_at,updated_at) VALUES (?,?,?,'Reusable draft','cover_letter','Retain me',now(),now())`, writingID, owner.ID, app.ID).Error)
	outside := ResearchTask{UserID: owner.ID, Title: "Independent", TaskType: "scholarship_research", Status: "draft"}
	must(db.Create(&outside).Error)
	shared := ApplicationProposal{UserID: owner.ID, ResearchTaskID: outside.ID, ApplicationID: &app.ID, Name: "Shared", Status: "approved", Priority: "high"}
	must(db.Create(&shared).Error)
	blocked, err := repo.PreviewResearchDeletion(ctx, &task)
	must(err)
	if len(blocked.Blockers) == 0 {
		t.Fatal("shared application must block deletion")
	}
	if err := repo.CascadeDeleteResearch(ctx, &task, blocked.ConfirmationToken); err == nil {
		t.Fatal("shared application deleted")
	}
	must(db.Delete(&shared).Error)
	must(db.Unscoped().Delete(&outside).Error)
	preview, err = repo.PreviewResearchDeletion(ctx, &task)
	must(err)
	if len(preview.Applications) != 1 || len(preview.ArchivedProfiles) != 1 {
		t.Fatalf("bad preview: %+v", preview)
	}
	must(repo.CascadeDeleteResearch(ctx, &task, preview.ConfirmationToken))
	for _, model := range []any{&ResearchTask{}, &ApplicationProposal{}, &application.Application{}, &application.ResearchRun{}} {
		var count int64
		must(db.Unscoped().Model(model).Count(&count).Error)
		if count != 0 {
			t.Fatalf("remaining %T: %d", model, count)
		}
	}

	var writing struct {
		ApplicationID *uuid.UUID
		Content       string
	}
	must(db.Table("writing_samples").Where("id = ?", writingID).First(&writing).Error)
	if writing.ApplicationID != nil || writing.Content != "Retain me" {
		t.Fatal("reusable writing was not retained and detached")
	}
	var remaining profile.ApplicantProfile
	must(db.First(&remaining, "id = ?", parent.ID).Error)
	must(db.Unscoped().First(&remaining, "id = ?", parent.ID).Error)
	var archived profile.ApplicantProfile
	must(db.Unscoped().First(&archived, "id = ?", tailored.ID).Error)
	if !archived.DeletedAt.Valid {
		t.Fatal("application profile was not archived")
	}
}
