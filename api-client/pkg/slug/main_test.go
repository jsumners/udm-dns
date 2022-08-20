package slug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func  TestHostname(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("foo-bar", Hostname("foo bar"))
	assert.Equal("FooBar", Hostname("FooBar"))
	assert.Equal("foo-b-r", Hostname("foo bår"))
	assert.Equal("", Hostname(""))
}
