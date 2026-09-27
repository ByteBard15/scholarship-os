package writing

import (
	"time"

	"github.com/google/uuid"
)

type CreateSampleRequest struct {
	UserID          *uuid.UUID  `json:"userId"`
	ApplicationID   *uuid.UUID  `json:"applicationId"`
	Title           string      `json:"title" validate:"required"`
	DocumentType    string      `json:"documentType" validate:"required,oneof=personal_statement motivation_letter recommendation_letter statement_of_purpose essay cover_letter other"`
	Content         string      `json:"content" validate:"required"`
	Origin          string      `json:"origin" validate:"omitempty,oneof=user agent import"`
	Status          string      `json:"status" validate:"omitempty,oneof=source draft suggested approved archived"`
	Description     *string     `json:"description"`
	TagNames        []string    `json:"tagNames" validate:"max=30,dive,max=100"`
	SourceSampleIDs []uuid.UUID `json:"sourceSampleIds" validate:"max=20,dive,required"`
}

type UpdateSampleRequest struct {
	Title        *string   `json:"title" validate:"omitempty,min=1"`
	DocumentType *string   `json:"documentType" validate:"omitempty,oneof=personal_statement motivation_letter recommendation_letter statement_of_purpose essay cover_letter other"`
	Content      *string   `json:"content" validate:"omitempty,min=1"`
	Status       *string   `json:"status" validate:"omitempty,oneof=source draft suggested approved archived"`
	Description  *string   `json:"description"`
	TagNames     *[]string `json:"tagNames" validate:"omitempty,max=30,dive,max=100"`
}

type CreateTagRequest struct {
	UserID *uuid.UUID `json:"userId"`
	Name   string     `json:"name" validate:"required,max=100"`
}

type MatchSamplesRequest struct {
	UserID        *uuid.UUID `json:"userId"`
	Tags          []string   `json:"tags" validate:"required,min=1,max=30,dive,max=100"`
	DocumentType  *string    `json:"documentType" validate:"omitempty,oneof=personal_statement motivation_letter recommendation_letter statement_of_purpose essay cover_letter other"`
	ApplicationID *uuid.UUID `json:"applicationId"`
	Limit         int        `json:"limit" validate:"omitempty,min=1,max=4"`
}

type TagResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	SampleCount int64     `json:"sampleCount"`
}

type SampleResponse struct {
	ID              uuid.UUID     `json:"id"`
	UserID          uuid.UUID     `json:"userId"`
	ApplicationID   *uuid.UUID    `json:"applicationId,omitempty"`
	Title           string        `json:"title"`
	DocumentType    string        `json:"documentType"`
	Content         string        `json:"content"`
	Origin          string        `json:"origin"`
	Status          string        `json:"status"`
	Description     *string       `json:"description,omitempty"`
	CreatedBy       *string       `json:"createdBy,omitempty"`
	Tags            []TagResponse `json:"tags"`
	SourceSampleIDs []uuid.UUID   `json:"sourceSampleIds,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type SampleMatchResponse struct {
	SampleResponse
	MatchedTags []string `json:"matchedTags"`
	Confidence  float64  `json:"confidence"`
}
