package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamicUrl(t *testing.T) {
	DynamicUrl("abc")
}

func TestBPath(t *testing.T) {
	BPath(PathXRelationStat)
}

func TestBVIDUrl(t *testing.T) {
	BVIDUrl("bv123")
}

func TestAddUAOption(t *testing.T) {
	AddUAOption()
	AddReferOption("bilibili")
	AddReferOption()
}

func TestIsAccountGiven(t *testing.T) {
	assert.False(t, IsAccountGiven())
	SetAccount("username", "password")
	assert.True(t, IsAccountGiven())
}

func TestParseUid(t *testing.T) {
	uid, err := ParseUid("123")
	assert.Nil(t, err)
	assert.EqualValues(t, 123, uid)

	uid, err = ParseUid("UID:456")
	assert.Nil(t, err)
	assert.EqualValues(t, 456, uid)

	uid, err = ParseUid("uid:789")
	assert.Nil(t, err)
	assert.EqualValues(t, 789, uid)
}

func TestCookieInfo(t *testing.T) {
	SetAccount("a", "b")
	assert.True(t, IsVerifyGiven())
	SetAccount("", "")

	_, err := freshAccountCookieInfo()
	assert.NotNil(t, err)
}

func TestSetVerify(t *testing.T) {
	defer atomicVerifyInfo.Store(new(VerifyInfo))

	SetVerify("wrong", "wrong")
	assert.EqualValues(t, "wrong", GetVerifyBiliJct())
}
