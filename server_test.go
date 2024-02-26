package envite

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(":8081", nil)
	assert.NotNil(t, s)
	assert.Equal(t, ":8081", s.httpServer.Addr)

	// Assert addr validation
	s = NewServer("8081", nil)
	assert.NotNil(t, s)
	assert.Equal(t, ":8081", s.httpServer.Addr)
}
