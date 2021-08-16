package youtube

import jsoniter "github.com/json-iterator/go"

const Site = "youtube"
const VideoView = "https://www.youtube.com/watch?v="

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func VideoViewUrl(videoId string) string {
	return VideoView + videoId
}
