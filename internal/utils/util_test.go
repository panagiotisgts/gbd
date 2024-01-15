package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindAndReplace(t *testing.T) {
	cfg := map[string]any{
		"db": map[string]any{
			"host": "localhost",
			"port": 5432,
			"security": map[string]any{
				"ssl": map[string]any{
					"mode": "disable",
					"tls":  true,
				},
			},
		},
	}

	FindAndReplace([]string{"db", "host"}, "127.0.0.1", cfg)
	FindAndReplace([]string{"db", "security", "ssl", "tls"}, false, cfg)

	require.Equal(t, "127.0.0.1", cfg["db"].(map[string]any)["host"])
	require.Equal(t, false, cfg["db"].(map[string]any)["security"].(map[string]any)["ssl"].(map[string]any)["tls"])
}
