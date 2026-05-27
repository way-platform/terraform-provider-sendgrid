package client

import (
	"context"

	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
)

// GetTemplate retrieves a template by ID.
func (c *Client) GetTemplate(
	ctx context.Context,
	id string,
) (*templates.TransactionalTemplate, error) {
	var t templates.TransactionalTemplate
	if err := c.doJSON(ctx, "GET", "/v3/templates/"+id, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
