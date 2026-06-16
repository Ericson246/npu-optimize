package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheSetGet(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	err := c.Set("test-key", []byte("hello world"), time.Hour)
	require.NoError(t, err)

	data, ok := c.Get("test-key")
	assert.True(t, ok)
	assert.Equal(t, []byte("hello world"), data)
}

func TestCacheGet_Miss(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	data, ok := c.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, data)
}

func TestCacheGet_Expired(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	err := c.Set("expired-key", []byte("data"), -time.Hour)
	require.NoError(t, err)

	data, ok := c.Get("expired-key")
	assert.False(t, ok)
	assert.Nil(t, data)

	_, err = os.Stat(filepath.Join(dir, "expired-key.json"))
	assert.True(t, os.IsNotExist(err), "expired entry should be removed")
}

func TestCacheSet_Overwrite(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	require.NoError(t, c.Set("key", []byte("first"), time.Hour))
	require.NoError(t, c.Set("key", []byte("second"), time.Hour))

	data, ok := c.Get("key")
	assert.True(t, ok)
	assert.Equal(t, []byte("second"), data)
}

func TestHardwareFingerprint(t *testing.T) {
	fp1 := HardwareFingerprint("RTX 4060", 8192, 32768, 8)
	fp2 := HardwareFingerprint("RTX 4060", 8192, 32768, 8)
	fp3 := HardwareFingerprint("RTX 4090", 24576, 65536, 16)

	assert.Equal(t, fp1, fp2)
	assert.NotEqual(t, fp1, fp3)
	assert.Len(t, fp1, 64)
}

func TestFingerprint(t *testing.T) {
	fp1 := Fingerprint("hello")
	fp2 := Fingerprint("hello")
	fp3 := Fingerprint("world")

	assert.Equal(t, fp1, fp2)
	assert.NotEqual(t, fp1, fp3)
	assert.Len(t, fp1, 64)
}
