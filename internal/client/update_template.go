package client

import (
	"context"

	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
)

// UpdateTemplate updates a template's name.
func (c *Client) UpdateTemplate(
	ctx context.Context,
	id string,
	req templates.UpdateTemplateJSONRequestBody,
) (*templates.TransactionalTemplate, error) {
	var t templates.TransactionalTemplate
	if err := c.doJSON(ctx, "PATCH", "/v3/templates/"+id, req, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
