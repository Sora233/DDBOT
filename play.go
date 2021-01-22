package main

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/asmcos/requests"
	"regexp"
	"strings"
	"time"
)

type R struct {
	Sub []*gabs.Container
}

func (r *R) search(key string, j *gabs.Container) {
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

func play() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req := requests.RequestsWithContext(ctx)
	req.Proxy("http://172.16.1.135:3128")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")
	req.Header.Set("accept-language", "zh-CN")
	resp, err := req.Get("https://www.youtube.com/channel/UCUKngXhjnKJ6KCyuC7ejI_w/videos?view=57&flow=grid")
	if err != nil {
		panic(err)
	}
	defer resp.R.Body.Close()
	content := resp.Content()
	var reg *regexp.Regexp
	if strings.Contains(string(content), `window["ytInitialData"]`) {
		reg = regexp.MustCompile("window\\[\"ytInitialData\"\\] = (?P<json>.*);")
	} else {
		reg = regexp.MustCompile(">var ytInitialData = (?P<json>.*?);</script>")
	}
	result := reg.FindSubmatch(content)

	j, err := gabs.ParseJSON(result[reg.SubexpIndex("json")])
	if err != nil {
		panic(err)
	}
	r := new(R)
	r.search("gridVideoRenderer", j)
	r.search("videoRenderer", j)

	for _, s := range r.Sub {
		q := s.Search("thumbnailOverlays", "0",
			"thumbnailOverlayTimeStatusRenderer", "text",
			"accessibility", "accessibilityData", "label")
		if q != nil {
			fmt.Println(q.String())
		} else {
			//fmt.Println(s.String())
		}
	}

}
