package client

import "context"

// GetImage retrieves image metadata by ID.
func (c *Client) GetImage(ctx context.Context, id string) (*Image, error) {
	var img Image
	if err := c.doJSON(ctx, "GET", "/v3/images/"+id, nil, &img); err != nil {
		return nil, err
	}
	return &img, nil
}
