package twitcasting

import "github.com/nobuf/cas"

type LiveStatus struct {
	UserId string
	Living bool
	Movie  *cas.MovieContainer
}

func (tc *TwitCastConcern) GetIsLive(user string) (*LiveStatus, error) {

	status := &LiveStatus{
		UserId: user,
		Living: true, // 一开始为正在直播，直到 404 错误出现
	}

	movie, err := tc.client.UserCurrentLive(user)

	if err != nil {
		// 没有直播
		if tcErr, ok := err.(*cas.RequestError); ok && tcErr.Content.Code == 404 {
			status.Living = false
		} else {
			return nil, err
		}
	}

	// 如果出现错误, movie 将会是 nil, 可参考 cas 源码
	status.Movie = movie
	return status, nil
}
