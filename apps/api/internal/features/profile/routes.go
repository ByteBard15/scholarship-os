package profile

import "github.com/go-chi/chi/v5"

func MountRoutes(r chi.Router, h *Handler) {
	r.Get("/users/{userID}/profiles", h.List)
	r.Post("/users/{userID}/profiles", h.Create)
	r.Get("/profiles/{profileID}", h.Get)
	r.Patch("/profiles/{profileID}", h.Update)
	r.Get("/profiles/{profileID}/full", h.GetFull)
	r.Get("/profiles/{profileID}/effective", h.GetEffective)
	r.Get("/profiles/{profileID}/lineage", h.GetLineage)
	r.Get("/profiles/{profileID}/completeness", h.GetCompleteness)
	r.Get("/profiles/{profileID}/compare/{otherProfileID}", h.Compare)
	r.Get("/profiles/{profileID}/overrides", h.ListOverrides)
	r.Post("/profiles/{profileID}/overrides", h.CreateOverride)
	r.Patch("/profiles/{profileID}/overrides/{overrideID}", h.UpdateOverride)
	r.Delete("/profiles/{profileID}/overrides/{overrideID}", h.DeleteOverride)
	r.Get("/profiles/{profileID}/snapshots", h.ListSnapshots)
	r.Post("/profiles/{profileID}/snapshots", h.CreateSnapshot)
	r.Get("/profiles/{profileID}/snapshots/{snapshotID}", h.GetSnapshot)
	r.Get("/profiles/{profileID}/documents", h.ListDocuments)
	r.Post("/profiles/{profileID}/documents", h.UploadDocument)
	r.Get("/profiles/{profileID}/documents/{documentID}", h.GetDocument)
	r.Delete("/profiles/{profileID}/documents/{documentID}", h.DeleteDocument)
	r.Get("/profiles/{profileID}/imports", h.ListImports)
	r.Post("/profiles/{profileID}/imports", h.CreateImport)
	r.Get("/profiles/{profileID}/imports/{importID}", h.GetImport)
	r.Get("/profiles/{profileID}/imports/{importID}/candidates", h.ListImportCandidates)
	r.Patch("/profiles/{profileID}/imports/{importID}/candidates/{candidateID}", h.ReviewCandidate)
	r.Post("/profiles/{profileID}/imports/{importID}/apply", h.ApplyImport)
	r.Get("/profiles/{profileID}/evidence", h.ListEvidence)
	r.Post("/profiles/{profileID}/evidence", h.CreateEvidence)
	r.Get("/profiles/{profileID}/personal-info", h.GetPersonalInfo)
	r.Put("/profiles/{profileID}/personal-info", h.PutPersonalInfo)
	r.Patch("/profiles/{profileID}/personal-info", h.PatchPersonalInfo)

	for _, entry := range []struct {
		path string
		kind SectionKind
	}{{"education", EducationKind}, {"employment", EmploymentKind}, {"projects", ProjectsKind}, {"publications", PublicationsKind}, {"articles", ArticlesKind}, {"skills", SkillsKind}, {"research-interests", ResearchInterestsKind}, {"career-goals", CareerGoalsKind}, {"certifications", CertificationsKind}, {"awards", AwardsKind}, {"volunteering", VolunteeringKind}} {
		base := "/profiles/{profileID}/" + entry.path
		r.Get(base, h.ListSection(entry.kind))
		r.Post(base, h.CreateSection(entry.kind))
		r.Get(base+"/{sectionID}", h.GetSection(entry.kind))
		r.Patch(base+"/{sectionID}", h.UpdateSection(entry.kind))
		r.Delete(base+"/{sectionID}", h.DeleteSection(entry.kind))
	}
}
