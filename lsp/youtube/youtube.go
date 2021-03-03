package youtube

const Site = "youtube"
const VideoView = "https://www.youtube.com/watch?v="

func VideoViewUrl(videoId string) string {
	return VideoView + videoId
}
