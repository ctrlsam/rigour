package api

import (
	"context"

	"github.com/ctrlsam/rigour/pkg/types"
)

// CampaignController is the interface consumed by API handlers to manage
// campaign lifecycle. Both the in-process Orchestrator and any future
// distributed scheduler satisfy this interface.
type CampaignController interface {
	Start(ctx context.Context, c *types.Campaign)
	Cancel(id string) error
	Pause(id string) error
	IsActive(id string) bool
}
