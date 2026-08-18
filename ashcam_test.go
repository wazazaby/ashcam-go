package ashcam

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type recorder struct {
	req  *http.Request
	body string
	// status is answered instead of a 200 when set.
	status int
}

func newTestClient(t *testing.T, body string) (*Client, *recorder) {
	t.Helper()

	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		rec.req, rec.body = r, string(data)
		if rec.status != 0 {
			w.WriteHeader(rec.status)
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	return NewClient(WithBaseURL(srv.URL), WithCredentials("user", "pass")), rec
}

func TestRequestPaths(t *testing.T) {
	start := time.Unix(1700000000, 0)
	end := time.Unix(1750000000, 0)

	tests := []struct {
		name   string
		call   func(*Client) error
		method string
		target string
	}{{
		name:   "images",
		call:   func(c *Client) error { _, err := c.GetImages(t.Context(), "redoubt-2"); return err },
		method: http.MethodGet,
		target: "/imageApi/webcam/redoubt-2",
	}, {
		name: "images by days old",
		call: func(c *Client) error {
			_, err := c.GetImages(t.Context(), "redoubt-2", DaysOld(7), Limit(10), OldestImageFirst())
			return err
		},
		method: http.MethodGet,
		target: "/imageApi/webcam/redoubt-2/7/oldestFirst/10",
	}, {
		name: "images by time range",
		call: func(c *Client) error {
			_, err := c.GetImages(t.Context(), "redoubt-2", TimeRange(start, end))
			return err
		},
		method: http.MethodGet,
		target: "/imageApi/webcam/redoubt-2/1700000000/1750000000/newestFirst/0",
	}, {
		name:   "recent images",
		call:   func(c *Client) error { _, err := c.GetRecentImages(t.Context(), 5); return err },
		method: http.MethodGet,
		target: "/imageApi/recent/5",
	}, {
		name:   "interesting images",
		call:   func(c *Client) error { _, err := c.GetInterestingImages(t.Context(), 0); return err },
		method: http.MethodGet,
		target: "/imageApi/interesting",
	}, {
		name:   "interesting images by days old",
		call:   func(c *Client) error { _, err := c.GetInterestingImages(t.Context(), 3); return err },
		method: http.MethodGet,
		target: "/imageApi/interesting/3",
	}, {
		name:   "uninteresting images",
		call:   func(c *Client) error { _, err := c.GetUninterestingImages(t.Context(), 3); return err },
		method: http.MethodGet,
		target: "/imageApi/uninteresting/3",
	}, {
		name: "set interesting code",
		call: func(c *Client) error {
			_, err := c.SetInterestingCode(t.Context(), "42", VolcanicActivity)
			return err
		},
		method: http.MethodPut,
		target: "/imageApi/interestingCode/42/V",
	}, {
		name:   "webcam",
		call:   func(c *Client) error { _, err := c.GetWebcam(t.Context(), "akunIsland-N"); return err },
		method: http.MethodGet,
		target: "/webcamApi/webcam/akunIsland-N",
	}, {
		name:   "webcams",
		call:   func(c *Client) error { _, err := c.GetWebcams(t.Context()); return err },
		method: http.MethodGet,
		target: "/webcamApi/webcams",
	}, {
		name: "webcams geojson",
		call: func(c *Client) error {
			_, err := c.GetWebcamsGeoJSON(t.Context(), GeoArea{Lat1: 60, Lat2: 70, Long1: -150, Long2: -160})
			return err
		},
		method: http.MethodGet,
		target: "/webcamApi/geojson?lat1=60&lat2=70&long1=-150&long2=-160",
	}, {
		name:   "refresh all webcams",
		call:   func(c *Client) error { _, err := c.RefreshAllWebcams(t.Context()); return err },
		method: http.MethodPost,
		target: "/webcamApi/refreshAll",
	}, {
		name: "assign clear image",
		call: func(c *Client) error {
			_, err := c.AssignClearImage(t.Context(), "redoubt-2", "42")
			return err
		},
		method: http.MethodGet,
		target: "/webcamApi/assignClearImage/redoubt-2/42",
	}, {
		name:   "housekeep",
		call:   func(c *Client) error { _, err := c.Housekeep(t.Context()); return err },
		method: http.MethodGet,
		target: "/adminApi/housekeep",
	}, {
		name:   "purge images",
		call:   func(c *Client) error { _, err := c.PurgeImages(t.Context(), 30); return err },
		method: http.MethodGet,
		target: "/adminApi/purgeImages/30",
	}, {
		name:   "mirrors",
		call:   func(c *Client) error { _, err := c.GetMirrors(t.Context()); return err },
		method: http.MethodGet,
		target: "/adminApi/mirrors",
	}, {
		name:   "auth check",
		call:   func(c *Client) error { return c.AuthCheck(t.Context()) },
		method: http.MethodGet,
		target: "/authcheck",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, rec := newTestClient(t, "{}")
			require.NoError(t, test.call(client))

			require.Equal(t, test.method, rec.req.Method)
			require.Equal(t, test.target, rec.req.URL.RequestURI())
			require.Equal(t, "user", rec.req.Header.Get("username"))
			require.Equal(t, "pass", rec.req.Header.Get("password"))
		})
	}
}

func TestRequestParameterErrors(t *testing.T) {
	client, _ := newTestClient(t, "{}")

	_, err := client.GetImages(t.Context(), "redoubt-2", DaysOld(7), TimeRange(time.Now(), time.Now()))
	require.ErrorIs(t, err, ErrDaysOldAndTimeRangeCantBeUsedTogether)

	_, err = client.GetUninterestingImages(t.Context(), 0)
	require.ErrorIs(t, err, ErrDaysOldRequired)
}

func TestErrorStatuses(t *testing.T) {
	for status, target := range map[int]error{
		http.StatusForbidden:           ErrNotAuthorized,
		http.StatusUnauthorized:        ErrNotAuthorized,
		http.StatusNotFound:            ErrNotFound,
		http.StatusInternalServerError: nil,
	} {
		client, rec := newTestClient(t, "nope")
		rec.status = status

		_, err := client.GetWebcams(t.Context())

		apiErr, ok := errors.AsType[*APIError](err)
		require.True(t, ok)
		require.Equal(t, status, apiErr.StatusCode)
		require.Equal(t, "nope", apiErr.Body)

		if target != nil {
			require.ErrorIs(t, err, target)
		}
	}
}

// The error path only keeps the first KB of the body, so it has to drain the
// rest for the connection to be reused.
func TestErrorPathReusesConnection(t *testing.T) {
	var connections atomic.Int64

	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, strings.Repeat("x", 5000), http.StatusInternalServerError)
	}))
	srv.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	srv.Start()
	defer srv.Close()

	client := NewClient(WithBaseURL(srv.URL))
	for range 5 {
		_, err := client.GetWebcams(t.Context())

		apiErr, ok := errors.AsType[*APIError](err)
		require.True(t, ok)
		require.Len(t, apiErr.Body, 1024)
	}

	require.Equal(t, int64(1), connections.Load())
}

func TestUploadImage(t *testing.T) {
	client, rec := newTestClient(t, `{"status":"ok"}`)

	res, err := client.UploadImage(t.Context(), ImageUpload{
		Image:           strings.NewReader("jpeg-data"),
		WebcamCode:      "redoubt-2",
		Timestamp:       time.Unix(1700000000, 0),
		InterestingCode: VolcanicActivity,
		NotNewest:       true,
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"status":"ok"}`, string(res))

	require.Equal(t, http.MethodPost, rec.req.Method)
	require.Equal(t, "/imageApi/uploadImage", rec.req.URL.Path)
	require.Contains(t, rec.req.Header.Get("Content-Type"), "multipart/form-data")

	for _, want := range []string{
		`name="webcamCode"`, "redoubt-2",
		`name="imageTimestamp"`, "1700000000",
		`name="interestingCode"`, "V",
		`name="notNewest"`, "Y",
		`name="file"; filename="image.jpg"`, "jpeg-data",
	} {
		require.Contains(t, rec.body, want)
	}
}

func TestCreateOrUpdateWebcam(t *testing.T) {
	client, rec := newTestClient(t, "{}")

	_, err := client.CreateOrUpdateWebcam(t.Context(), WebcamInput{
		Code:       "TEST",
		Name:       "Test webcam label",
		Latitude:   90,
		Longitude:  160,
		Elevation:  900,
		BearingDeg: 45,
		IsFAA:      false,
	})
	require.NoError(t, err)

	require.Equal(t, "application/json", rec.req.Header.Get("Content-Type"))
	require.JSONEq(t, `{
		"webcamCode":"TEST","webcamName":"Test webcam label","latitude":"90","longitude":"160",
		"elevationM":"900","bearingDeg":"45","externalUrl":"","vnum":"","vName":"","faaInd":"N"
	}`, rec.body)
}

func TestAssignVolcano(t *testing.T) {
	client, rec := newTestClient(t, "{}")

	_, err := client.AssignVolcano(t.Context(), "TEST", "311320", "Akutan")
	require.NoError(t, err)

	require.Equal(t, "/webcamApi/assignVolcano", rec.req.URL.Path)
	require.JSONEq(t, `{"webcamCode":"TEST","vnum":"311320","vName":"Akutan"}`, rec.body)
}

func TestDecodeGeoJSON(t *testing.T) {
	const payload = `{
		"type":"FeatureCollection",
		"metadata":{"generated":1787055156,"title":"USGS Webcams For Defined Region","count":1,"params":"lat1=60"},
		"features":[{
			"type":"Feature",
			"properties":{
				"webcamName":"Iliamna","webcamCode":"iliamna_nnl","elevationM":40,"bearingDeg":264,
				"externalUrl":null,"hasImages":"N","imageTotal":0,"newestImage":[],
				"firstImageDate":"Tue, 18 Aug 2026 12:12:36 +0000","firstImageTimestamp":0,
				"lastImageDate":"Tue, 18 Aug 2026 12:12:36 +0000","lastImageTimestamp":0,
				"clearImageUrl":"https://example.org/clear.jpg","vnum":"313020","vName":"Iliamna"
			},
			"geometry":{"type":"Point","coordinates":[-151,60]},
			"id":1053
		}]
	}`

	var res GeoJSONResponse
	require.NoError(t, json.Unmarshal([]byte(payload), &res))

	require.Equal(t, "FeatureCollection", res.Type)
	require.Equal(t, 1, res.Metadata.Count)
	require.Len(t, res.Features, 1)

	feature := res.Features[0]
	require.Equal(t, 1053, feature.ID)
	require.Equal(t, []float64{-151, 60}, feature.Geometry.Coordinates)
	require.Equal(t, "iliamna_nnl", feature.Properties.Code)
	require.Equal(t, 313020, feature.Properties.VNum)
	require.Equal(t, StateNo, feature.Properties.HasImages)
	require.Zero(t, feature.Properties.NewestImage.ID)
}

func TestDecodeImages(t *testing.T) {
	// The API omits or nulls out the sun informations and the meta timestamps
	// when it has none to report.
	const payload = `{
		"images":[{
			"imageId":48002367,"md5":"74ab8e96","webcamCode":"kilauea-kw-cam","newestForWebcam":"Y",
			"imageTimestamp":1787055121,"imageDate":"Tue, 18 Aug 2026 12:12:01 +0000",
			"isNighttimeInd":"Y","interestingCode":"V","imageUrl":"https://example.org/i.jpg",
			"suninfo":{"timezone":"UTC","time_in_unixtime":1787055121,"time_in":"Tue, 18 Aug 2026 12:12:01 +0000",
				"civil_twilight_sunrise":"Tue, 18 Aug 2026 15:40:08 +0000","civil_twilight_sunrise_unixtime":1787067608,
				"civil_twilight_sunset":"Wed, 19 Aug 2026 05:09:41 +0000","civil_twilight_sunset_unixtime":1787116181}
		},{
			"imageId":48002365,"isNighttimeInd":"?","interestingCode":"U","suninfo":null
		}],
		"meta":{"imageTotal":2,"firstImageTimestamp":null,"lastImageTimestamp":null,
			"apiUrl":"https://example.org/api","querySec":11}
	}`

	client, _ := newTestClient(t, payload)

	res, err := client.GetRecentImages(t.Context(), 2)
	require.NoError(t, err)
	require.Len(t, res.Images, 2)
	require.Equal(t, 2, res.Meta.ImageTotal)
	require.Zero(t, res.Meta.FirstImageTimestamp)

	first := res.Images[0]
	require.Equal(t, VolcanicActivity, first.InterestingCode)
	require.True(t, first.InterestingCode.IsInteresting())
	require.Equal(t, StateYes, first.IsNightTime)
	require.Equal(t, int64(1787055121), first.Date.Time().Unix())
	require.Equal(t, "UTC", first.SunInformations.Timezone)

	second := res.Images[1]
	require.Equal(t, UnknownVolcanicActivity, second.InterestingCode)
	require.Equal(t, StateUnknown, second.IsNightTime)
	require.Zero(t, second.SunInformations.Timezone)
}

func TestLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test hitting the live API")
	}

	client := NewClient()

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	webcam, err := client.GetWebcam(ctx, "akunIsland-N")
	require.NoError(t, err)
	require.Equal(t, "akunIsland-N", webcam.Webcam.Code)

	webcams, err := client.GetWebcams(ctx)
	require.NoError(t, err)
	require.Len(t, webcams.Webcams, webcams.WebcamsMeta.Total)

	images, err := client.GetImages(ctx, "akunIsland-N", DaysOld(7), Limit(2))
	require.NoError(t, err)
	require.LessOrEqual(t, len(images.Images), 2)

	recent, err := client.GetRecentImages(ctx, 2)
	require.NoError(t, err)
	require.Len(t, recent.Images, 2)

	geojson, err := client.GetWebcamsGeoJSON(ctx, GeoArea{Lat1: 60, Lat2: 70, Long1: -150, Long2: -160})
	require.NoError(t, err)
	require.Len(t, geojson.Features, geojson.Metadata.Count)

	_, err = client.GetUninterestingImages(ctx, 1)
	require.NoError(t, err)

	// Every write endpoint needs credentials we don't have here.
	require.ErrorIs(t, client.AuthCheck(ctx), ErrNotAuthorized)
}
