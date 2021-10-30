package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFeedList(t *testing.T) {
	assert.NotNil(t, FeedPageOpt(1))
	assert.NotNil(t, FeedPageSizeOpt(1))
	_, err := FeedList()
	assert.NotNil(t, err)
}
