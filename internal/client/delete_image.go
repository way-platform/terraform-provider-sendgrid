package client

import "context"

// DeleteImage deletes an image.
func (c *Client) DeleteImage(ctx context.Context, id string) error {
	return c.doJSON(ctx, "DELETE", "/v3/images/"+id, nil, nil)
}
