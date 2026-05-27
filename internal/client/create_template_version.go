package client

import (
	"context"
	"fmt"

	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
)

// CreateTemplateVersion creates a new version for a template.
func (c *Client) CreateTemplateVersion(
	ctx context.Context,
	templateID string,
	req templates.CreateTemplateVersionJSONRequestBody,
) (*templates.TransactionalTemplateVersionOutput, error) {
	path := fmt.Sprintf("/v3/templates/%s/versions", templateID)
	var tv templates.TransactionalTemplateVersionOutput
	if err := c.doJSON(ctx, "POST", path, req, &tv); err != nil {
		return nil, err
	}
	return &tv, nil
}
