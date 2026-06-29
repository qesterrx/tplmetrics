package aes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAES(t *testing.T) {

	keyAES, err := GenAESKey()
	assert.NoError(t, err)

	tests := []struct {
		name        string
		text        string
		key         []byte
		expectError bool
	}{
		{
			name:        "encrypt empty message",
			text:        "",
			key:         keyAES,
			expectError: false,
		},
		{
			name:        "short data",
			text:        "hello world",
			key:         keyAES,
			expectError: false,
		},
		{
			name:        "large data",
			text:        strings.Repeat("a", 1024*1024*1024),
			key:         keyAES,
			expectError: false,
		},
		{
			name:        "wrong key",
			text:        strings.Repeat("a", 1024*1024*1024),
			key:         []byte("invalid"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptGCM([]byte(tt.text), tt.key)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEqual(t, []byte(tt.text), encrypted)

			decrypted, err := DecryptGCM(encrypted, tt.key)
			assert.NoError(t, err)
			assert.NotEqual(t, encrypted, decrypted)
			assert.Equal(t, tt.text, string(decrypted))
		})
	}

}
