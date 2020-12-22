package moderate

const (
	RatingIndexEveryone = iota + 1
	RatingIndexTeen
	RatingIndexAdult
)

type ModerateTag map[string]float64

type InappropriateContentResponse struct {
	UrlClassified string `json:"url_classified"`
	RatingIndex   int    `json:"rating_index"`
	RatingLetter  string `json:"rating_letter"`
	Predictions   struct {
		Teen     float64 `json:"teen"`
		Everyone float64 `json:"everyone"`
		Adult    float64 `json:"adult"`
	} `json:"predictions"`
	RatingLabel string `json:"rating_label"`
	ErrorCode   int    `json:"error_code"`
	Error       string `json:"error"`
}

type ModerateResponse struct {
	InappropriateContentResponse
}

type AnimeResponse struct {
	InappropriateContentResponse
	Tags      ModerateTag `json:"tags"`
	Character ModerateTag `json:"character"`
	Copyright ModerateTag `json:"copyright"`
	Error     string      `json:"error"`
}
