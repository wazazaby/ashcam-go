package ashcam

// TODO(teddy): add json struct tags
type Webcam struct {
	Code           string
	Name           string
	Latitude       float64
	Longitude      float64
	Elevation      float64
	BearingDegrees int
	ExternalURL    string

	VNum  string // TODO(teddy): use numeric string type ?
	VName string

	HasImages  Bool
	ImageTotal int

	FirstImageDate     DateRFC1123Z
	FistImageTimestamp int
	LastImageDate      DateRFC1123Z
	LastImageTimestamp int

	ClearImageURL         string
	CurrentImageURL       string
	CurrentMediumImageURL string
	CurrentThumbImageURL  string

	NewestImage Image

	Timezone string // useful when sun informations are not set

	SunInformations SunInformations
}
