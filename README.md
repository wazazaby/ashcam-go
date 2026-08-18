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

A non 2xx response gives an `*ashcam.APIError`, which matches
`ashcam.ErrNotAuthorized` on 401/403 and `ashcam.ErrNotFound` on 404 - note the
API answers 500, not 404, for an unknown webcam code or image identifier:

```go
if apiErr, ok := errors.AsType[*ashcam.APIError](err); ok {
    log.Printf("%s failed with %d: %s", apiErr.URL, apiErr.StatusCode, apiErr.Body)
}
```

The `imageApi/webcam` query parameter variant is not implemented: the API
returns a 500 for it.
