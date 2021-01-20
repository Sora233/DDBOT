package aliyun

import "errors"

var ErrNotInit = errors.New("not init")

type AuditResponse struct {
	Data struct {
		Results []struct {
			Code       int    `json:"Code"`
			Message    string `json:"Message"`
			DataId     string `json:"DataId"`
			ImageURL   string `json:"ImageURL"`
			TaskId     string `json:"TaskId"`
			SubResults []struct {
				Label      string  `json:"Label"`
				Scene      string  `json:"Scene"`
				Rate       float64 `json:"Rate"`
				Suggestion string  `json:"Suggestion"`
			}
		} `json:"Results"`
	} `json:"Data"`
}
