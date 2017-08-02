package striphtml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripTags(t *testing.T) {
	assert.Equal(t, "This is some text", StripTags("<span class='some class'>This is <br>some text</span>"))
}
