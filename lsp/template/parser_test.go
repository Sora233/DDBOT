package template

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParser(t *testing.T) {
	p := new(Parser)

	testCases := []string{
		`hello {{ .pic a/b/c }} yolo`,
		`this is {{score}} yes`,
		`wrong template {{ }}, no`,
		`wrong template }} no`,
		`wrong template {{ .asd }} no`,
	}

	expected := [][]INode{
		{
			&stringNode{s: `hello `},
			&picNode{uri: `a/b/c`},
			&stringNode{s: ` yolo`},
		},
		{
			&stringNode{s: `this is `},
			&keyNode{key: `score`},
			&stringNode{s: ` yes`},
		},
		nil,
		nil,
		nil,
	}

	assertNode := func(t *testing.T, node1, node2 INode) {
		if node1 == nil {
			assert.Nil(t, node2)
			return
		}
		switch s1 := node1.(type) {
		case *stringNode:
			s2, ok := node2.(*stringNode)
			assert.True(t, ok)
			assert.EqualValues(t, s1.s, s2.s)
		case *picNode:
			s2, ok := node2.(*picNode)
			assert.True(t, ok)
			assert.EqualValues(t, s1.uri, s2.uri)
		case *keyNode:
			s2, ok := node2.(*keyNode)
			assert.True(t, ok)
			assert.EqualValues(t, s1.key, s2.key)
		}
	}

	assert.EqualValues(t, len(expected), len(testCases))
	for idx := range expected {
		c := testCases[idx]
		e := expected[idx]
		err := p.Parse(c)
		if e == nil {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		for cnt := range e {
			assert.True(t, p.Next())
			assertNode(t, e[cnt], p.Peek())
		}
	}
}
