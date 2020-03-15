package mmh3

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestExpectedValues(t *testing.T) {
	t.Run("Hash32", func(t *testing.T) {
		readExpectedValues(t, "testdata/32/expected.txt", func(key, value string) {
			expectedValue, err := strconv.ParseUint(value, 10, 0)
			require.NoError(t, err)

			assert.Equal(t, uint32(expectedValue), Hash32([]byte(key)))
		})
	})

	t.Run("Hash128", func(t *testing.T) {
		readExpectedValues(t, "testdata/128/expected.txt", func(key, value string) {
			expectedValue, err := hex.DecodeString(value)
			require.NoError(t, err)

			assert.Equal(t, expectedValue, Hash128x64([]byte(key)))
		})
	})
}

func readExpectedValues(t *testing.T, filename string, lineCB func(key, value string)) {
	f, err := os.Open(filename)
	require.NoError(t, err)

	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ",", 2)
		require.Len(t, parts, 2)
		lineCB(parts[0], parts[1])
	}

	require.NoError(t, scanner.Err())
}

func TestAll(t *testing.T) {
	s := []byte("hello")
	if Hash32(s) != 0x248bfa47 {
		t.Fatal("32bit hello")
	}
	if fmt.Sprintf("%x", Hash128(s)) != "029bbd41b3a7d8cb191dae486a901e5b" {
		t.Fatal("128bit hello")
	}
	s = []byte("Winter is coming")
	if Hash32(s) != 0x43617e8f {
		t.Fatal("32bit winter")
	}
	if fmt.Sprintf("%x", Hash128(s)) != "95eddc615d3b376c13fb0b0cead849c5" {
		t.Fatal("128bit winter")
	}
}

// Test the x64 optimized hash against Hash128
func TestOptimizedX64(t *testing.T) {
	keys := []string{
		"hello",
		"Winter is coming",
	}

	for _, k := range keys {
		h128 := Hash128([]byte(k))
		h128x64 := Hash128x64([]byte(k))

		if string(h128) != string(h128x64) {
			t.Fatalf("Expected same hashes for %s, but got %x and %x", k, h128, h128x64)
		}
	}

}

func TestHashWriter128(t *testing.T) {
	s := []byte("hello")
	h := HashWriter128{}
	h.Write(s)
	res := h.Sum(nil)
	if fmt.Sprintf("%x", res) != "029bbd41b3a7d8cb191dae486a901e5b" {
		t.Fatal("128bit hello")
	}
	s = []byte("Winter is coming")
	h.Reset()
	h.Write(s)
	res = h.Sum(nil)
	if fmt.Sprintf("%x", res) != "95eddc615d3b376c13fb0b0cead849c5" {
		t.Fatal("128bit hello")
	}
	str := "Winter is coming"
	h.Reset()
	h.WriteString(str)
	res = h.Sum(nil)
	if fmt.Sprintf("%x", res) != "95eddc615d3b376c13fb0b0cead849c5" {
		t.Fatal("128bit hello")
	}

}
