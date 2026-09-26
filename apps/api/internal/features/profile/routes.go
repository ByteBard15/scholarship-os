package profile

import (
	"github.com/byte/scholarship-os/apps/api/internal/features/ownership"
	"github.com/go-chi/chi/v5"
)

func MountRoutes(r chi.Router, h *Handler, owners *ownership.Middleware) {
	r.With(owners.User("list user profiles")).Get("/users/{userID}/profiles", h.List)
	r.With(owners.User("create user profile")).Post("/users/{userID}/profiles", h.Create)
	r.With(owners.Profile("get profile")).Get("/profiles/{profileID}", h.Get)
	r.With(owners.Profile("update profile")).Patch("/profiles/{profileID}", h.Update)
	r.With(owners.Profile("get full profile")).Get("/profiles/{profileID}/full", h.GetFull)
	r.With(owners.Profile("resolve effective profile")).Get("/profiles/{profileID}/effective", h.GetEffective)
	r.With(owners.Profile("get profile lineage")).Get("/profiles/{profileID}/lineage", h.GetLineage)
	r.With(owners.Profile("get profile completeness")).Get("/profiles/{profileID}/completeness", h.GetCompleteness)
	r.With(owners.Profile("compare source profile"), owners.ProfileParam("compare target profile", "otherProfileID")).Get("/profiles/{profileID}/compare/{otherProfileID}", h.Compare)
	r.With(owners.Profile("list profile overrides")).Get("/profiles/{profileID}/overrides", h.ListOverrides)
	r.With(owners.Profile("create profile override")).Post("/profiles/{profileID}/overrides", h.CreateOverride)
	r.With(owners.Profile("update profile override")).Patch("/profiles/{profileID}/overrides/{overrideID}", h.UpdateOverride)
	r.With(owners.Profile("delete profile override")).Delete("/profiles/{profileID}/overrides/{overrideID}", h.DeleteOverride)
	r.With(owners.Profile("list profile snapshots")).Get("/profiles/{profileID}/snapshots", h.ListSnapshots)
	r.With(owners.Profile("create profile snapshot")).Post("/profiles/{profileID}/snapshots", h.CreateSnapshot)
	r.With(owners.Profile("get profile snapshot")).Get("/profiles/{profileID}/snapshots/{snapshotID}", h.GetSnapshot)
	r.With(owners.Profile("list profile documents")).Get("/profiles/{profileID}/documents", h.ListDocuments)
	r.With(owners.Profile("upload profile document")).Post("/profiles/{profileID}/documents", h.UploadDocument)
	r.With(owners.Profile("get profile document")).Get("/profiles/{profileID}/documents/{documentID}", h.GetDocument)
	r.With(owners.Profile("delete profile document")).Delete("/profiles/{profileID}/documents/{documentID}", h.DeleteDocument)
	r.With(owners.Profile("list profile imports")).Get("/profiles/{profileID}/imports", h.ListImports)
	r.With(owners.Profile("create profile import")).Post("/profiles/{profileID}/imports", h.CreateImport)
	r.With(owners.Profile("get profile import")).Get("/profiles/{profileID}/imports/{importID}", h.GetImport)
	r.With(owners.Profile("list profile import candidates")).Get("/profiles/{profileID}/imports/{importID}/candidates", h.ListImportCandidates)
	r.With(owners.Profile("review profile import candidate")).Patch("/profiles/{profileID}/imports/{importID}/candidates/{candidateID}", h.ReviewCandidate)
	r.With(owners.Profile("apply profile import")).Post("/profiles/{profileID}/imports/{importID}/apply", h.ApplyImport)
	r.With(owners.Profile("list profile evidence")).Get("/profiles/{profileID}/evidence", h.ListEvidence)
	r.With(owners.Profile("create profile evidence")).Post("/profiles/{profileID}/evidence", h.CreateEvidence)
	r.With(owners.Profile("get profile personal information")).Get("/profiles/{profileID}/personal-info", h.GetPersonalInfo)
	r.With(owners.Profile("replace profile personal information")).Put("/profiles/{profileID}/personal-info", h.PutPersonalInfo)
	r.With(owners.Profile("update profile personal information")).Patch("/profiles/{profileID}/personal-info", h.PatchPersonalInfo)

	for _, entry := range []struct {
		path string
		kind SectionKind
	}{{"education", EducationKind}, {"employment", EmploymentKind}, {"projects", ProjectsKind}, {"publications", PublicationsKind}, {"articles", ArticlesKind}, {"skills", SkillsKind}, {"research-interests", ResearchInterestsKind}, {"career-goals", CareerGoalsKind}, {"certifications", CertificationsKind}, {"awards", AwardsKind}, {"volunteering", VolunteeringKind}} {
		base := "/profiles/{profileID}/" + entry.path
		r.With(owners.Profile("list "+entry.path)).Get(base, h.ListSection(entry.kind))
		r.With(owners.Profile("create "+entry.path+" entry")).Post(base, h.CreateSection(entry.kind))
		r.With(owners.Profile("get "+entry.path+" entry")).Get(base+"/{sectionID}", h.GetSection(entry.kind))
		r.With(owners.Profile("update "+entry.path+" entry")).Patch(base+"/{sectionID}", h.UpdateSection(entry.kind))
		r.With(owners.Profile("delete "+entry.path+" entry")).Delete(base+"/{sectionID}", h.DeleteSection(entry.kind))
	}
}
