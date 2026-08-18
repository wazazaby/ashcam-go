package ashcam

import (
	"context"
	"net/http"
	"strconv"
)

// AuthCheck reports whether the configured credentials are accepted by the API.
func (c *Client) AuthCheck(ctx context.Context) error {
	_, err := c.raw(ctx, http.MethodGet, "/authcheck", nil)
	return err
}

// Housekeep removes the image records without image file and the image files
// without record, then refreshes the webcam statistics.
//
// Requires credentials.
func (c *Client) Housekeep(ctx context.Context) ([]byte, error) {
	return c.raw(ctx, http.MethodGet, "/adminApi/housekeep", nil)
}

// PurgeImages removes the images older than daysToKeep - except the ones
// flagged as interesting - then refreshes the webcam statistics.
//
// Requires credentials.
func (c *Client) PurgeImages(ctx context.Context, daysToKeep int) ([]byte, error) {
	return c.raw(ctx, http.MethodGet, "/adminApi/purgeImages/"+strconv.Itoa(daysToKeep), nil)
}

// GetMirrors returns the local mirrors information.
//
// Requires credentials.
func (c *Client) GetMirrors(ctx context.Context) ([]byte, error) {
	return c.raw(ctx, http.MethodGet, "/adminApi/mirrors", nil)
}
