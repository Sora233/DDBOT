package main

import (
	"github.com/Jeffail/gabs/v2"
	"github.com/Sora233/Sora233-MiraiGo/lsp/youtube"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/local_proxy_pool"
)

type R struct {
	Sub []*gabs.Container
}

func (r *R) search(key string, j *gabs.Container, limit int) {
	if len(j.ChildrenMap()) != 0 {
		for k, v := range j.ChildrenMap() {
			if k == key {
				r.Sub = append(r.Sub, v)
				if limit != -1 && len(r.Sub) >= limit {
					return
				}
				continue
			}
			r.search(key, v, limit)
			if limit != -1 && len(r.Sub) >= limit {
				return
			}
		}
	} else {
		for _, c := range j.Children() {
			if len(c.ChildrenMap()) != 0 {
				r.search(key, c, limit)
				if limit != -1 && len(r.Sub) >= limit {
					return
				}
			}
		}
	}
}

func play() {
	proxy_pool.Init(local_proxy_pool.NewLocalPool([]string{"172.16.1.135:3128"}))
	xv, err := youtube.XFetchInfo("UCflNPJUJ4VQh1hGDNK7bsFg")
	if err != nil {
		panic(err)
	}
	for _, v := range xv {
		if v.IsLive() {

		}
	}
}
