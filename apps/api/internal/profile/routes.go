package profile

import "github.com/go-chi/chi/v5"

func MountRoutes(r chi.Router, h *Handler) {
	r.Get("/users/{userID}/profiles", h.List)
	r.Post("/users/{userID}/profiles", h.Create)
	r.Get("/profiles/{profileID}", h.Get)
	r.Patch("/profiles/{profileID}", h.Update)
	r.Get("/profiles/{profileID}/full", h.GetFull)
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
