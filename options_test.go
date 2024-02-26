package envite

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	assert.Equal(t, LogLevelTrace.String(), "TRACE")
	assert.Equal(t, LogLevelDebug.String(), "DEBUG")
	assert.Equal(t, LogLevelInfo.String(), "INFO")
	assert.Equal(t, LogLevelError.String(), "ERROR")
	assert.Equal(t, LogLevelFatal.String(), "FATAL")
}
