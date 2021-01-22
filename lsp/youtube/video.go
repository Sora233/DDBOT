package youtube

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"strconv"
	"strings"
	"time"
)

const VideoPath = "https://www.youtube.com/channel/%s/videos?view=57&flow=grid"

type Searcher struct {
	VideoList []*gabs.Container
}

func (r *Searcher) search(key string, j *gabs.Container) {
	if len(j.ChildrenMap()) != 0 {
		for k, v := range j.ChildrenMap() {
			if k == key {
				r.VideoList = append(r.VideoList, v)
				continue
			}
			r.search(key, v)
		}
	} else {
		for _, c := range j.Children() {
			if len(c.ChildrenMap()) != 0 {
				r.search(key, c)
			}
		}
	}
}

func XFetchInfo(channelID string) ([]*VideoInfo, error) {
	log := logger.WithField("channel_id", channelID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		log.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	path := fmt.Sprintf(VideoPath, channelID)
	resp, err := requests.Get(ctx, path, nil, 3,
		requests.HeaderOption("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36"),
		requests.HeaderOption("accept-language", "zh-CN"),
	)
	if err != nil {
		return nil, err
	}
	var videoInfos []*VideoInfo
	root, err := gabs.ParseJSON(resp.Content())
	if err != nil {
		return nil, err
	}
	var searcher = new(Searcher)
	searcher.search("gridVideoRenderer", root)
	searcher.search("videoRenderer", root)
	for _, videoJson := range searcher.VideoList {
		var i = new(VideoInfo)
		i.VideoId = strings.Trim(videoJson.S("videoId").String(), `"`)
		if videoJson.ExistsP("title.simpleText") {
			i.VideoTitle = videoJson.Path("title.simpleText").String()
		} else if videoJson.ExistsP("title.runs") {
			sb := strings.Builder{}
			for _, c := range videoJson.Path("title.runs").Children() {
				sb.WriteString(c.String())
			}
			i.VideoTitle = sb.String()
		}

		switch videoJson.S("thumbnailOverlays", "0",
			"thumbnailOverlayTimeStatusRenderer", "text",
			"accessibility", "accessibilityData", "label").String() {
		case "PREMIERE", "首播", "プレミア":
			i.VideoType = VideoType_FirstLive
		case "LIVE", "直播", "ライブ":
			i.VideoType = VideoType_Live
		case "null":
			log.Error("null video type")
			continue
		default:
			i.VideoType = VideoType_Video
		}

		switch videoJson.S("thumbnailOverlays", "0", "thumbnailOverlayTimeStatusRenderer", "style").String() {
		case "UPCOMING":
			i.VideoStatus = VideoStatus_Waiting
			i.VideoTimestamp, _ = strconv.ParseInt(videoJson.Path("upcomingEventData.upcomingEventData").String(), 10, 64)
		case "LIVE":
			i.VideoStatus = VideoStatus_Living
		case "null":
			log.Error("null video status")
			continue
		default:
			i.VideoStatus = VideoStatus_Upload
		}
		i.ChannelId = channelID
		videoInfos = append(videoInfos, i)
	}
	return videoInfos, nil
}
