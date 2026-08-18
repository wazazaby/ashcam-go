package ashcam

import (
	"context"
	"net/http"
	"strconv"
)

// AuthCheck reports whether the API accepts the credentials set with
// [WithCredentials]. It returns an error wrapping [ErrNotAuthorized] when it
// doesn't.
func (c *Client) AuthCheck(ctx context.Context) error {
	_, err := c.raw(ctx, http.MethodGet, "/authcheck", nil)
	return err
}

// Housekeep removes the image records without image file and the image files
// without record, then refreshes the webcam statistics.
//
// Requires [WithCredentials]. The API doesn't document a response payload, so
// the body is returned as is.
func (c *Client) Housekeep(ctx context.Context) ([]byte, error) {
	return c.raw(ctx, http.MethodGet, "/adminApi/housekeep", nil)
}

// PurgeImages removes the images older than daysToKeep - except the ones flagged
// with [VolcanicActivity], see [Client.SetInterestingCode] - then refreshes the
// webcam statistics.
//
// Requires [WithCredentials]. The API doesn't document a response payload, so
// the body is returned as is.
func (c *Client) PurgeImages(ctx context.Context, daysToKeep int) ([]byte, error) {
	return c.raw(ctx, http.MethodGet, "/adminApi/purgeImages/"+strconv.Itoa(daysToKeep), nil)
}

// GetMirrors returns the local mirrors information, which is what
// [WithBaseURL] expects a URL from.
//
// Requires [WithCredentials]. The API doesn't document a response payload, so
// the body is returned as is.
func (c *Client) GetMirrors(ctx context.Context) ([]byte, error) {
	return c.raw(ctx, http.MethodGet, "/adminApi/mirrors", nil)
}
