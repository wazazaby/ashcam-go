package ashcam

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTeddy(t *testing.T) {
	client := NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.GetWebcam(ctx, code)
	require.NoError(t, err)

	fmt.Println(res.Webcam.ClearImageURL)

	res2, err := client.GetWebcams(ctx)
	require.NoError(t, err)

	fmt.Println(res2.WebcamsMeta.Total)
}

var code = "akunIsland-N"
