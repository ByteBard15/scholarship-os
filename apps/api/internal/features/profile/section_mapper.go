package profile

import "github.com/google/uuid"

func newSection(k SectionKind, pid uuid.UUID, r SectionRequest) any {
	var v any
	switch k {
	case EducationKind:
		v = &EducationHistory{ProfileID: pid}
	case EmploymentKind:
		v = &EmploymentHistory{ProfileID: pid}
	case ProjectsKind:
		v = &Project{ProfileID: pid}
	case PublicationsKind:
		v = &Publication{ProfileID: pid}
	case ArticlesKind:
		v = &Article{ProfileID: pid}
	case SkillsKind:
		v = &Skill{ProfileID: pid}
	case ResearchInterestsKind:
		v = &ResearchInterest{ProfileID: pid}
	case CareerGoalsKind:
		v = &CareerGoal{ProfileID: pid}
	case CertificationsKind:
		v = &Certification{ProfileID: pid}
	case AwardsKind:
		v = &Award{ProfileID: pid}
	case VolunteeringKind:
		v = &VolunteerExperience{ProfileID: pid}
	}
	applySection(v, r)
	return v
}
func applySection(v any, r SectionRequest) {
	switch x := v.(type) {
	case *EducationHistory:
		if r.Institution != nil {
			x.Institution = *r.Institution
		}
		if r.Degree != nil {
			x.Degree = *r.Degree
		}
		if r.FieldOfStudy != nil {
			x.FieldOfStudy = *r.FieldOfStudy
		}
		x.StartDate = pickTime(r.StartDate, x.StartDate)
		x.EndDate = pickTime(r.EndDate, x.EndDate)
		if r.IsCurrent != nil {
			x.IsCurrent = *r.IsCurrent
		}
		x.Grade = pick(r.Grade, x.Grade)
		x.GradeScale = pick(r.GradeScale, x.GradeScale)
		x.Classification = pick(r.Classification, x.Classification)
		x.City = pick(r.City, x.City)
		x.Country = pick(r.Country, x.Country)
		x.Description = pick(r.Description, x.Description)
	case *EmploymentHistory:
		if r.Organization != nil {
			x.Organization = *r.Organization
		}
		if r.JobTitle != nil {
			x.JobTitle = *r.JobTitle
		}
		x.EmploymentType = pick(r.EmploymentType, x.EmploymentType)
		x.StartDate = pickTime(r.StartDate, x.StartDate)
		x.EndDate = pickTime(r.EndDate, x.EndDate)
		if r.IsCurrent != nil {
			x.IsCurrent = *r.IsCurrent
		}
		x.City = pick(r.City, x.City)
		x.Country = pick(r.Country, x.Country)
		x.Description = pick(r.Description, x.Description)
	case *Project:
		if r.Title != nil {
			x.Title = *r.Title
		}
		x.Organization = pick(r.Organization, x.Organization)
		x.StartDate = pickTime(r.StartDate, x.StartDate)
		x.EndDate = pickTime(r.EndDate, x.EndDate)
		x.Description = pick(r.Description, x.Description)
		x.URL = pick(r.URL, x.URL)
		x.RepositoryURL = pick(r.RepositoryURL, x.RepositoryURL)
	case *Publication:
		if r.Title != nil {
			x.Title = *r.Title
		}
		x.PublicationType = pick(r.PublicationType, x.PublicationType)
		x.Publisher = pick(r.Publisher, x.Publisher)
		x.PublicationDate = pickTime(r.PublicationDate, x.PublicationDate)
		x.DOI = pick(r.DOI, x.DOI)
		x.URL = pick(r.URL, x.URL)
		x.Description = pick(r.Description, x.Description)
	case *Article:
		if r.Title != nil {
			x.Title = *r.Title
		}
		x.PublicationDate = pickTime(r.PublicationDate, x.PublicationDate)
		x.URL = pick(r.URL, x.URL)
		x.Description = pick(r.Description, x.Description)
	case *Skill:
		if r.Name != nil {
			x.Name = *r.Name
		}
		x.Category = pick(r.Category, x.Category)
		x.Proficiency = pick(r.Proficiency, x.Proficiency)
		if r.YearsExperience != nil {
			x.YearsExperience = r.YearsExperience
		}
	case *ResearchInterest:
		if r.Name != nil {
			x.Name = *r.Name
		}
		x.Description = pick(r.Description, x.Description)
		if r.Priority != nil {
			x.Priority = r.Priority
		}
	case *CareerGoal:
		if r.Title != nil {
			x.Title = *r.Title
		}
		if r.Description != nil {
			x.Description = *r.Description
		}
		x.GoalType = pick(r.GoalType, x.GoalType)
		if r.Priority != nil {
			x.Priority = r.Priority
		}
	case *Certification:
		if r.Name != nil {
			x.Name = *r.Name
		}
		x.Issuer = pick(r.Issuer, x.Issuer)
		x.IssueDate = pickTime(r.IssueDate, x.IssueDate)
		x.ExpirationDate = pickTime(r.ExpirationDate, x.ExpirationDate)
		x.CredentialID = pick(r.CredentialID, x.CredentialID)
		x.URL = pick(r.URL, x.URL)
	case *Award:
		if r.Title != nil {
			x.Title = *r.Title
		}
		x.Issuer = pick(r.Issuer, x.Issuer)
		x.AwardDate = pickTime(r.AwardDate, x.AwardDate)
		x.Description = pick(r.Description, x.Description)
	case *VolunteerExperience:
		if r.Organization != nil {
			x.Organization = *r.Organization
		}
		if r.Role != nil {
			x.Role = *r.Role
		}
		x.StartDate = pickTime(r.StartDate, x.StartDate)
		x.EndDate = pickTime(r.EndDate, x.EndDate)
		if r.IsCurrent != nil {
			x.IsCurrent = *r.IsCurrent
		}
		x.Description = pick(r.Description, x.Description)
	}
}
func pick[T any](next, current *T) *T {
	if next != nil {
		return next
	}
	return current
}
func pickTime[T any](next, current *T) *T { return pick(next, current) }

func personalResponse(v *ProfilePersonalInfo) PersonalInfoResponse {
	return PersonalInfoResponse{ID: v.ID, ProfileID: v.ProfileID, FirstName: v.FirstName, MiddleName: v.MiddleName, LastName: v.LastName, PreferredName: v.PreferredName, Phone: v.Phone, City: v.City, StateOrRegion: v.StateOrRegion, Country: v.Country, Nationality: v.Nationality, LinkedInURL: v.LinkedInURL, GitHubURL: v.GitHubURL, WebsiteURL: v.WebsiteURL, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func sectionResponse(v any) SectionResponse {
	var o SectionResponse
	switch x := v.(type) {
	case EducationHistory:
		return sectionResponse(&x)
	case EmploymentHistory:
		return sectionResponse(&x)
	case Project:
		return sectionResponse(&x)
	case Publication:
		return sectionResponse(&x)
	case Article:
		return sectionResponse(&x)
	case Skill:
		return sectionResponse(&x)
	case ResearchInterest:
		return sectionResponse(&x)
	case CareerGoal:
		return sectionResponse(&x)
	case Certification:
		return sectionResponse(&x)
	case Award:
		return sectionResponse(&x)
	case VolunteerExperience:
		return sectionResponse(&x)
	case *EducationHistory:
		o = baseResponse(x.Base, x.ProfileID)
		o.Institution = &x.Institution
		o.Degree = &x.Degree
		o.FieldOfStudy = &x.FieldOfStudy
		o.StartDate = x.StartDate
		o.EndDate = x.EndDate
		o.IsCurrent = &x.IsCurrent
		o.Grade = x.Grade
		o.GradeScale = x.GradeScale
		o.Classification = x.Classification
		o.City = x.City
		o.Country = x.Country
		o.Description = x.Description
	case *EmploymentHistory:
		o = baseResponse(x.Base, x.ProfileID)
		o.Organization = &x.Organization
		o.JobTitle = &x.JobTitle
		o.EmploymentType = x.EmploymentType
		o.StartDate = x.StartDate
		o.EndDate = x.EndDate
		o.IsCurrent = &x.IsCurrent
		o.City = x.City
		o.Country = x.Country
		o.Description = x.Description
	case *Project:
		o = baseResponse(x.Base, x.ProfileID)
		o.Title = &x.Title
		o.Organization = x.Organization
		o.StartDate = x.StartDate
		o.EndDate = x.EndDate
		o.Description = x.Description
		o.URL = x.URL
		o.RepositoryURL = x.RepositoryURL
	case *Publication:
		o = baseResponse(x.Base, x.ProfileID)
		o.Title = &x.Title
		o.PublicationType = x.PublicationType
		o.Publisher = x.Publisher
		o.PublicationDate = x.PublicationDate
		o.DOI = x.DOI
		o.URL = x.URL
		o.Description = x.Description
	case *Article:
		o = baseResponse(x.Base, x.ProfileID)
		o.Title = &x.Title
		o.PublicationDate = x.PublicationDate
		o.URL = x.URL
		o.Description = x.Description
	case *Skill:
		o = baseResponse(x.Base, x.ProfileID)
		o.Name = &x.Name
		o.Category = x.Category
		o.Proficiency = x.Proficiency
		o.YearsExperience = x.YearsExperience
	case *ResearchInterest:
		o = baseResponse(x.Base, x.ProfileID)
		o.Name = &x.Name
		o.Description = x.Description
		o.Priority = x.Priority
	case *CareerGoal:
		o = baseResponse(x.Base, x.ProfileID)
		o.Title = &x.Title
		o.Description = &x.Description
		o.GoalType = x.GoalType
		o.Priority = x.Priority
	case *Certification:
		o = baseResponse(x.Base, x.ProfileID)
		o.Name = &x.Name
		o.Issuer = x.Issuer
		o.IssueDate = x.IssueDate
		o.ExpirationDate = x.ExpirationDate
		o.CredentialID = x.CredentialID
		o.URL = x.URL
	case *Award:
		o = baseResponse(x.Base, x.ProfileID)
		o.Title = &x.Title
		o.Issuer = x.Issuer
		o.AwardDate = x.AwardDate
		o.Description = x.Description
	case *VolunteerExperience:
		o = baseResponse(x.Base, x.ProfileID)
		o.Organization = &x.Organization
		o.Role = &x.Role
		o.StartDate = x.StartDate
		o.EndDate = x.EndDate
		o.IsCurrent = &x.IsCurrent
		o.Description = x.Description
	}
	return o
}
func baseResponse(b Base, pid uuid.UUID) SectionResponse {
	return SectionResponse{ID: b.ID, ProfileID: pid, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt}
}
func sectionResponses(v any) []SectionResponse {
	out := []SectionResponse{}
	switch rows := v.(type) {
	case []EducationHistory:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []EmploymentHistory:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []Project:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []Publication:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []Article:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []Skill:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []ResearchInterest:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []CareerGoal:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []Certification:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []Award:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	case []VolunteerExperience:
		for _, x := range rows {
			out = append(out, sectionResponse(x))
		}
	}
	return out
}
func fullResponse(p *ApplicantProfile) FullProfileResponse {
	o := FullProfileResponse{ProfileResponse: profileResponse(p), Education: sectionResponses(p.Education), Employment: sectionResponses(p.Employment), Projects: sectionResponses(p.Projects), Publications: sectionResponses(p.Publications), Articles: sectionResponses(p.Articles), Skills: sectionResponses(p.Skills), ResearchInterests: sectionResponses(p.ResearchInterests), CareerGoals: sectionResponses(p.CareerGoals), Certifications: sectionResponses(p.Certifications), Awards: sectionResponses(p.Awards), Volunteering: sectionResponses(p.Volunteering)}
	if p.PersonalInfo != nil {
		x := personalResponse(p.PersonalInfo)
		o.PersonalInfo = &x
	}
	return o
}
