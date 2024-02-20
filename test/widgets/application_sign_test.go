package test

import (
	"testing"

	"github.com/dnwsilver/tld/internal/pkg/widgets"

	"github.com/stretchr/testify/assert"
)

func TestApplicationSign(t *testing.T) {
	widget := widgets.ApplicationSign("John Travolta", 1994)

	result := widget.Render()

	assert.EqualValues(t, "\uF1F9 John Travolta, 1994", result)
}
