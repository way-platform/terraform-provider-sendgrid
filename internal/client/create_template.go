package client

import (
	"context"

	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
)

// CreateTemplate creates a new dynamic template.
func (c *Client) CreateTemplate(
	ctx context.Context,
	req templates.CreateTemplateJSONRequestBody,
) (*templates.TransactionalTemplate, error) {
	var t templates.TransactionalTemplate
	if err := c.doJSON(ctx, "POST", "/v3/templates", req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
