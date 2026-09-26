package catalog

import "github.com/go-chi/chi/v5"

func MountRoutes(r chi.Router, h *Handler) {
	r.Get("/institutions", h.ListInstitutions)
	r.Post("/institutions", h.CreateInstitution)
	r.Get("/institutions/{institutionID}", h.GetInstitution)
	r.Patch("/institutions/{institutionID}", h.UpdateInstitution)
	r.Get("/programmes", h.ListProgrammes)
	r.Post("/programmes", h.CreateProgramme)
	r.Get("/programmes/{programmeID}", h.GetProgramme)
	r.Patch("/programmes/{programmeID}", h.UpdateProgramme)
	r.Get("/scholarships", h.ListScholarships)
	r.Post("/scholarships", h.CreateScholarship)
	r.Get("/scholarships/{scholarshipID}", h.GetScholarship)
	r.Patch("/scholarships/{scholarshipID}", h.UpdateScholarship)
}
