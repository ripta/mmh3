package mmh3

import (
	"bufio"

	"encoding/binary"
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
	if fmt.Sprintf("%x", Hash128(s).Bytes()) != "029bbd41b3a7d8cb191dae486a901e5b" {
		t.Fatal("128bit hello")
	}
	s = []byte("Winter is coming")
	if Hash32(s) != 0x43617e8f {
		t.Fatal("32bit winter")
	}
	if fmt.Sprintf("%x", Hash128(s).Bytes()) != "95eddc615d3b376c13fb0b0cead849c5" {
		t.Fatal("128bit winter")
	}
}

func Test128Values(t *testing.T) {
	key := []byte("hello world")

	h := Hash128(key)

	h1, h2 := h.Values()

	b := h.Bytes()

	out := make([]byte, 16)
	h.Write(out)

	assert.Equal(t, h1, binary.LittleEndian.Uint64(b[0:]))
	assert.Equal(t, h2, binary.LittleEndian.Uint64(b[8:]))

	assert.Equal(t, h1, binary.LittleEndian.Uint64(out[0:]))
	assert.Equal(t, h2, binary.LittleEndian.Uint64(out[8:]))

	assert.Equal(t, b, Hash128x64(key))

	WriteHash128x64(key, out)
	assert.Equal(t, b, out)
}

func TestHashWriter128(t *testing.T) {
	s := []byte("hello")
	h := HashWriter128{}
	_, _ = h.Write(s)
	res := h.Sum(nil)
	if fmt.Sprintf("%x", res) != "029bbd41b3a7d8cb191dae486a901e5b" {
		t.Fatal("128bit hello")
	}
	s = []byte("Winter is coming")
	h.Reset()
	_, _ = h.Write(s)
	res = h.Sum(nil)
	if fmt.Sprintf("%x", res) != "95eddc615d3b376c13fb0b0cead849c5" {
		t.Fatal("128bit hello")
	}
	str := "Winter is coming"
	h.Reset()
	_, _ = h.WriteString(str)
	res = h.Sum(nil)
	if fmt.Sprintf("%x", res) != "95eddc615d3b376c13fb0b0cead849c5" {
		t.Fatal("128bit hello")
	}

}
