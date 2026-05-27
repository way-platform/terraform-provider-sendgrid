package client

import (
	"context"
	"fmt"

	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
)

// GetTemplateVersion retrieves a specific template version.
func (c *Client) GetTemplateVersion(
	ctx context.Context,
	templateID, versionID string,
) (*templates.TransactionalTemplateVersionOutput, error) {
	path := fmt.Sprintf("/v3/templates/%s/versions/%s", templateID, versionID)
	var tv templates.TransactionalTemplateVersionOutput
	if err := c.doJSON(ctx, "GET", path, nil, &tv); err != nil {
		return nil, err
	}
	return &tv, nil
}
