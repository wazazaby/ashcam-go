# ashcam-go

Go client for the USGS ASHCAM API -> https://avo-volcview.wr.usgs.gov/ashcam-api/.

```go
client := ashcam.NewClient()

webcams, err := client.GetWebcams(ctx)
images, err := client.GetImages(ctx, "akunIsland-N", ashcam.DaysOld(7), ashcam.Limit(10))
```

Every write and admin endpoint needs credentials:

```go
client := ashcam.NewClient(ashcam.WithCredentials(username, password))

_, err := client.UploadImage(ctx, ashcam.ImageUpload{
    WebcamCode: "akunIsland-N",
    Timestamp:  time.Now(),
    Image:      file,
})
```

Those endpoints have no documented response payload, so they return the raw body
as `[]byte`.

The API runs on two instances, and they are not interchangeable. Both serve the
same webcam catalog, but each has its own database: `DefaultBaseURL` (the
default) keeps the full image archive, while `AVOBaseURL` - the Alaska Volcano
Observatory one - only keeps a deep archive for the Alaska and Yellowstone
webcams, and its image IDs are unrelated to the other instance's.

```go
client := ashcam.NewClient(ashcam.WithBaseURL(ashcam.AVOBaseURL))
```

A non 2xx response gives an `*ashcam.APIError`, which matches
`ashcam.ErrNotAuthorized` on 401/403 and `ashcam.ErrNotFound` on 404 - note the
API answers 500, not 404, for an unknown webcam code or image identifier:

```go
if apiErr, ok := errors.AsType[*ashcam.APIError](err); ok {
    log.Printf("%s failed with %d: %s", apiErr.URL, apiErr.StatusCode, apiErr.Body)
}
```

Two API quirks worth knowing, both verified against the live instances:

- the `imageApi/webcam` query parameter variant is not implemented here, the API
  returns a 500 for it
- `OldestImageFirst` and a non-zero `Limit` can't be combined, the API returns a
  500 for that pair
