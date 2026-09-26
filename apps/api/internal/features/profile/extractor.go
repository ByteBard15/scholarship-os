package profile

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrTextExtractionUnavailable = errors.New("text extraction is not available for this document format")

type TextExtractor interface {
	Extract(context.Context, *ProfileDocument, io.Reader) (string, error)
}
type DocumentTextExtractor struct{ maxBytes int64 }

func NewDocumentTextExtractor(maxBytes int64) *DocumentTextExtractor {
	return &DocumentTextExtractor{maxBytes: maxBytes}
}
func (e *DocumentTextExtractor) Extract(_ context.Context, document *ProfileDocument, reader io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, e.maxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > e.maxBytes {
		return "", ErrInvalidDocument
	}
	switch document.MimeType {
	case "text/plain":
		return string(data), nil
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return extractDOCX(data)
	case "application/pdf":
		return "", fmt.Errorf("%w: PDF extraction is deferred; upload TXT or DOCX for import", ErrTextExtractionUnavailable)
	default:
		return "", ErrInvalidDocument
	}
}
func extractDOCX(data []byte) (string, error) {
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open DOCX: %w", err)
	}
	for _, file := range archive.File {
		if file.Name != "word/document.xml" {
			continue
		}
		stream, err := file.Open()
		if err != nil {
			return "", err
		}
		defer stream.Close()
		decoder := xml.NewDecoder(stream)
		var parts []string
		for {
			token, err := decoder.Token()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return "", fmt.Errorf("parse DOCX: %w", err)
			}
			start, ok := token.(xml.StartElement)
			if !ok || start.Name.Local != "t" {
				continue
			}
			var value string
			if err = decoder.DecodeElement(&value, &start); err != nil {
				return "", err
			}
			if strings.TrimSpace(value) != "" {
				parts = append(parts, value)
			}
		}
		return strings.Join(parts, " "), nil
	}
	return "", errors.New("DOCX does not contain word/document.xml")
}

type ExtractedFact[T any] struct {
	Data       T        `json:"data"`
	Confidence *float64 `json:"confidence,omitempty"`
	SourceText *string  `json:"sourceText,omitempty"`
}
type ExtractedPersonalInfo struct {
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
}
type ExtractedEducation struct {
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"fieldOfStudy"`
}
type ExtractedEmployment struct {
	Organization string  `json:"organization"`
	JobTitle     string  `json:"jobTitle"`
	Description  *string `json:"description,omitempty"`
}
type ExtractedProject struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}
type ExtractedPublication struct {
	Title string `json:"title"`
}
type ExtractedArticle struct {
	Title string `json:"title"`
}
type ExtractedSkill struct {
	Name     string  `json:"name"`
	Category *string `json:"category,omitempty"`
}
type ExtractedCertification struct {
	Name   string  `json:"name"`
	Issuer *string `json:"issuer,omitempty"`
}
type ExtractedAward struct {
	Title  string  `json:"title"`
	Issuer *string `json:"issuer,omitempty"`
}
type ExtractedVolunteerExperience struct {
	Organization string `json:"organization"`
	Role         string `json:"role"`
}
type ExtractedProfile struct {
	PersonalInfo   *ExtractedFact[ExtractedPersonalInfo]         `json:"personalInfo,omitempty"`
	Education      []ExtractedFact[ExtractedEducation]           `json:"education"`
	Employment     []ExtractedFact[ExtractedEmployment]          `json:"employment"`
	Projects       []ExtractedFact[ExtractedProject]             `json:"projects"`
	Publications   []ExtractedFact[ExtractedPublication]         `json:"publications"`
	Articles       []ExtractedFact[ExtractedArticle]             `json:"articles"`
	Skills         []ExtractedFact[ExtractedSkill]               `json:"skills"`
	Certifications []ExtractedFact[ExtractedCertification]       `json:"certifications"`
	Awards         []ExtractedFact[ExtractedAward]               `json:"awards"`
	Volunteering   []ExtractedFact[ExtractedVolunteerExperience] `json:"volunteering"`
}

type ProfileExtractor interface {
	Extract(context.Context, string) (*ExtractedProfile, error)
	Provider() string
	Model() string
}
type DeterministicExtractor struct{}

func NewDeterministicExtractor() *DeterministicExtractor { return &DeterministicExtractor{} }
func (*DeterministicExtractor) Provider() string         { return "deterministic" }
func (*DeterministicExtractor) Model() string            { return "labeled-lines-v1" }
func (*DeterministicExtractor) Extract(_ context.Context, text string) (*ExtractedProfile, error) {
	result := &ExtractedProfile{}
	confidence := 0.99
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		label, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		values := splitValues(value)
		source := line
		switch strings.ToUpper(strings.TrimSpace(label)) {
		case "PERSONAL":
			if len(values) >= 2 {
				result.PersonalInfo = &ExtractedFact[ExtractedPersonalInfo]{Data: ExtractedPersonalInfo{FirstName: values[0], LastName: values[1]}, Confidence: &confidence, SourceText: &source}
			}
		case "SKILL", "SKILLS":
			for _, name := range strings.Split(value, ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					result.Skills = append(result.Skills, ExtractedFact[ExtractedSkill]{Data: ExtractedSkill{Name: name}, Confidence: &confidence, SourceText: &source})
				}
			}
		case "EDUCATION":
			if len(values) >= 3 {
				result.Education = append(result.Education, ExtractedFact[ExtractedEducation]{Data: ExtractedEducation{Institution: values[0], Degree: values[1], FieldOfStudy: values[2]}, Confidence: &confidence, SourceText: &source})
			}
		case "EMPLOYMENT":
			if len(values) >= 2 {
				result.Employment = append(result.Employment, ExtractedFact[ExtractedEmployment]{Data: ExtractedEmployment{Organization: values[0], JobTitle: values[1]}, Confidence: &confidence, SourceText: &source})
			}
		case "PROJECT":
			if len(values) >= 1 {
				result.Projects = append(result.Projects, ExtractedFact[ExtractedProject]{Data: ExtractedProject{Title: values[0]}, Confidence: &confidence, SourceText: &source})
			}
		case "PUBLICATION":
			if len(values) >= 1 {
				result.Publications = append(result.Publications, ExtractedFact[ExtractedPublication]{Data: ExtractedPublication{Title: values[0]}, Confidence: &confidence, SourceText: &source})
			}
		case "ARTICLE":
			if len(values) >= 1 {
				result.Articles = append(result.Articles, ExtractedFact[ExtractedArticle]{Data: ExtractedArticle{Title: values[0]}, Confidence: &confidence, SourceText: &source})
			}
		case "CERTIFICATION":
			if len(values) >= 1 {
				result.Certifications = append(result.Certifications, ExtractedFact[ExtractedCertification]{Data: ExtractedCertification{Name: values[0]}, Confidence: &confidence, SourceText: &source})
			}
		case "AWARD":
			if len(values) >= 1 {
				result.Awards = append(result.Awards, ExtractedFact[ExtractedAward]{Data: ExtractedAward{Title: values[0]}, Confidence: &confidence, SourceText: &source})
			}
		case "VOLUNTEERING":
			if len(values) >= 2 {
				result.Volunteering = append(result.Volunteering, ExtractedFact[ExtractedVolunteerExperience]{Data: ExtractedVolunteerExperience{Organization: values[0], Role: values[1]}, Confidence: &confidence, SourceText: &source})
			}
		}
	}
	return result, nil
}
func splitValues(value string) []string {
	raw := strings.Split(value, "|")
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
