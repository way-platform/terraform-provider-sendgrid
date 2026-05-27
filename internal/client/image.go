package client

// Image represents a SendGrid hosted image (no OAS spec available).
type Image struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
