package ownership

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrResourceNotFound = errors.New("owned resource not found")

// Repository resolves workspace owners at the HTTP authorization boundary.
// Domain repositories still load the complete records used by services.
type Repository interface {
	ProfileOwner(context.Context, uuid.UUID) (uuid.UUID, error)
	ApplicationOwner(context.Context, uuid.UUID) (uuid.UUID, error)
	ApplicationProposalOwner(context.Context, uuid.UUID) (uuid.UUID, error)
	QuestionnaireOwner(context.Context, uuid.UUID) (uuid.UUID, error)
	QuestionOwner(context.Context, uuid.UUID) (uuid.UUID, error)
	InformationRequestOwner(context.Context, uuid.UUID) (uuid.UUID, error)
	ResearchTaskOwner(context.Context, uuid.UUID) (uuid.UUID, error)
}
