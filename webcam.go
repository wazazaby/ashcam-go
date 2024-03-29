package ashcam

// TODO(teddy): add json struct tags
type Webcam struct {
	FirstImageDate        DateRFC1123Z
	LastImageDate         DateRFC1123Z
	ExternalURL           string
	Timezone              string // useful when sun informations are not set
	Code                  string
	ClearImageURL         string
	VNum                  string // TODO(teddy): use numeric string type ?
	VName                 string
	CurrentThumbImageURL  string
	CurrentMediumImageURL string
	CurrentImageURL       string
	Name                  string
	SunInformations       SunInformations
	NewestImage           Image
	BearingDegrees        int
	LastImageTimestamp    int
	FistImageTimestamp    int
	ImageTotal            int
	Elevation             float64
	Longitude             float64
	Latitude              float64
	HasImages             Bool
}
