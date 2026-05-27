package client

import (
	"context"
	"fmt"
)

// DeleteTemplateVersion deletes a template version.
func (c *Client) DeleteTemplateVersion(ctx context.Context, templateID, versionID string) error {
	path := fmt.Sprintf("/v3/templates/%s/versions/%s", templateID, versionID)
	return c.doJSON(ctx, "DELETE", path, nil, nil)
}
