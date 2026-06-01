package redeye

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewTimeMsg(t *testing.T) {
	now := time.Date(2025, time.March, 9, 10, 11, 12, 0, time.UTC)
	got := NewTimeMsg(now)

	require.NotNil(t, got)
	require.Equal(t, 2025, got.Year)
	require.Equal(t, time.March, got.Month)
	require.Equal(t, 9, got.Day)
	require.Equal(t, 10, got.Hour)
	require.Equal(t, 11, got.Minute)
	require.Equal(t, 12, got.Second)
	require.Equal(t, "setTime", got.Action)
}
