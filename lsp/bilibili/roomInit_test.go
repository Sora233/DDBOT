package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoomInit(t *testing.T) {
	const testRoom = 97505
	_, err := RoomInit(testRoom)
	assert.Nil(t, err)
}
