package youtube

import (
	"context"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const VideoPath = "https://www.youtube.com/channel/%s/videos?view=57&flow=grid"

type Searcher struct {
	Sub []*gabs.Container
}

func (r *Searcher) search(key string, j *gabs.Container) {
	if len(j.ChildrenMap()) != 0 {
		for k, v := range j.ChildrenMap() {
			if k == key {
				r.Sub = append(r.Sub, v)
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

// very sb
func XFetchInfo(channelID string) ([]*VideoInfo, error) {
	log := logger.WithField("channel_id", channelID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		log.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()

	var channelName string

	path := fmt.Sprintf(VideoPath, channelID)
	resp, err := requests.Get(ctx, path, nil, 3,
		requests.HeaderOption("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36"),
		requests.HeaderOption("accept-language", "zh-CN"),
		requests.ProxyOption(proxy_pool.PreferOversea),
	)
	if err != nil {
		return nil, err
	}
	content := resp.Content()
	var reg *regexp.Regexp
	if strings.Contains(string(content), `window["ytInitialData"]`) {
		reg = regexp.MustCompile("window\\[\"ytInitialData\"\\] = (?P<json>.*);")
	} else {
		reg = regexp.MustCompile(">var ytInitialData = (?P<json>.*?);</script>")
	}
	result := reg.FindSubmatch(content)

	if len(result) <= reg.SubexpIndex("json") {
		return nil, errors.New("no json data matched")
	}

	root, err := gabs.ParseJSON(result[reg.SubexpIndex("json")])
	if err != nil {
		return nil, err
	}

	var videoSearcher = new(Searcher)
	var infoSearcher = new(Searcher)

	videoSearcher.search("gridVideoRenderer", root)
	videoSearcher.search("videoRenderer", root)
	infoSearcher.search("channelMetadataRenderer", root)

	reg = regexp.MustCompile(`\\u[0-9]{4}`)
	if len(infoSearcher.Sub) >= 1 {
		channelName = infoSearcher.Sub[0].S("title").String()
		allCode := reg.FindAllString(channelName, -1)
		for _, code := range allCode {
			unquote, err := utils.UnquoteString(code)
			if err != nil {
				log.WithField("quote_string", code).Errorf("unquote failed %v, err", err)
				continue
			}
			channelName = strings.ReplaceAll(channelName, code, unquote)
		}
	} else {
		channelName = "<nil>"
	}
	channelName = strings.Trim(channelName, `"`)

	var idSet = make(map[string]bool)

	var videoInfos []*VideoInfo
	for _, videoJson := range videoSearcher.Sub {
		var i = new(VideoInfo)
		i.VideoId = strings.Trim(videoJson.S("videoId").String(), `"`)

		if _, found := idSet[i.VideoId]; !found {
			idSet[i.VideoId] = true
		} else {
			continue
		}

		if videoJson.ExistsP("title.simpleText") {
			i.VideoTitle = strings.Trim(videoJson.Path("title.simpleText").String(), `"`)
		} else if videoJson.ExistsP("title.runs") {
			sb := strings.Builder{}
			for _, c := range videoJson.Path("title.runs").Children() {
				sb.WriteString(strings.Trim(c.S("text").String(), `"`))
			}
			i.VideoTitle = sb.String()
		}

		switch strings.Trim(videoJson.S("thumbnailOverlays", "0",
			"thumbnailOverlayTimeStatusRenderer", "text",
			"accessibility", "accessibilityData", "label").String(), `"`) {
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

		switch strings.Trim(videoJson.S("thumbnailOverlays", "0", "thumbnailOverlayTimeStatusRenderer", "style").String(), `"`) {
		case "UPCOMING":
			i.VideoStatus = VideoStatus_Waiting
			i.VideoTimestamp, _ = strconv.ParseInt(strings.Trim(videoJson.Path("upcomingEventData.startTime").String(), `"`), 10, 64)
		case "LIVE":
			i.VideoStatus = VideoStatus_Living
		case "null":
			log.Error("null video status")
			continue
		default:
			i.VideoStatus = VideoStatus_Upload
		}
		i.ChannelId = channelID
		i.ChannelName = channelName

		var size int64 = -1
		for _, obj := range videoJson.S("thumbnail", "thumbnails").Children() {
			if obj.Exists("height") {
				height, err := strconv.ParseInt(obj.S("height").String(), 10, 64)
				if err != nil {
					continue
				}
				if size < height {
					size = height
					i.Cover = strings.Trim(obj.S("url").String(), `"`)
				}
			}
		}
		videoInfos = append(videoInfos, i)
	}
	log.WithField("video_count", len(videoInfos)).Tracef("fetch info")
	return videoInfos, nil
}
