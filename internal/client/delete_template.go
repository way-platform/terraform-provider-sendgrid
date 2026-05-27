package client

import "context"

// DeleteTemplate deletes a template.
func (c *Client) DeleteTemplate(ctx context.Context, id string) error {
	return c.doJSON(ctx, "DELETE", "/v3/templates/"+id, nil, nil)
}
