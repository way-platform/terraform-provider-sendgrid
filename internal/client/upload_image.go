package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// UploadImage uploads an image file to SendGrid.
func (c *Client) UploadImage(ctx context.Context, filePath string) (*Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on read-only file

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		part, err := writer.CreateFormFile("upload", filepath.Base(filePath))
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, f); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(writer.Close())
	}()

	resp, err := c.do(ctx, "POST", "/v3/images", pr, writer.FormDataContentType())
	if err != nil {
		pr.Close() //nolint:errcheck // unblock goroutine on request failure
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp)
	}

	var img Image
	if err := json.NewDecoder(resp.Body).Decode(&img); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &img, nil
}
