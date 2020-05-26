package mmh3

import (
	"bufio"
	"crypto/rand"
	"io"

	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("Hash128Writer", func(t *testing.T) {
		readExpectedValues(t, "testdata/128/expected.txt", func(key, value string) {
			expectedValue, err := hex.DecodeString(value)
			require.NoError(t, err)

			hw := HashWriter128{}
			_, _ = hw.Write([]byte(key))

			h := make([]byte, 16)

			hw.Sum(h[:0])

			assert.Equal(t, expectedValue, h)
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

// TestBoundaries forces every block/tail path to be exercised for Sum128.
// Borrowed from twmb/murmur3
func TestBoundaries(t *testing.T) {
	const maxCheck = 17
	var data [maxCheck]byte
	for i := 0; !t.Failed() && i < 20; i++ {
		// Check all zeros the first iteration.
		for size := 0; size <= maxCheck; size++ {
			test := data[:size]

			hw := HashWriter128{}
			_, _ = hw.Write(test)

			g128h1, g128h2 := hw.Sum128().Values()

			c128h1, c128h2 := Hash128(test).Values()

			if g128h1 != c128h1 {
				t.Errorf("size #%d: in: %x, g128h1 (%d) != c128h1 (%d); attempt #%d", size, test, g128h1, c128h1, i)
			}
			if g128h2 != c128h2 {
				t.Errorf("size #%d: in: %x, g128h2 (%d) != c128h2 (%d); attempt #%d", size, test, g128h2, c128h2, i)
			}
		}
		// Randomize the data for all subsequent tests.
		_, _ = io.ReadFull(rand.Reader, data[:])
	}
}

// Borrowed from twmb/murmur3
func TestIncremental(t *testing.T) {
	var data = []struct {
		h1 uint64
		h2 uint64
		s  string
	}{
		{0x0000000000000000, 0x0000000000000000, ""},
		{0xcbd8a7b341bd9b02, 0x5b1e906a48ae1d19, "hello"},
		{0x342fac623a5ebc8e, 0x4cdcbc079642414d, "hello, world"},
		{0xb89e5988b737affc, 0x664fc2950231b2cb, "19 Jan 2038 at 3:14:07 AM"},
		{0xcd99481f9ee902c9, 0x695da1a38987b6e7, "The quick brown fox jumps over the lazy dog."},
	}
	for _, elem := range data {
		hw := HashWriter128{}
		for i, j, k := 0, 0, len(elem.s); i < k; i = j {
			j = 2*i + 3
			if j > k {
				j = k
			}
			s := elem.s[i:j]
			_, _ = hw.Write([]byte(s))
		}

		if v1, v2 := hw.Sum128().Values(); v1 != elem.h1 || v2 != elem.h2 {
			t.Errorf("'%s': 0x%x-0x%x (want 0x%x-0x%x)", elem.s, v1, v2, elem.h1, elem.h2)
		}
	}
}

func Benchmark128Branches(b *testing.B) {
	for length := 0; length <= 16; length++ {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf := make([]byte, length)
			b.SetBytes(int64(length))
			b.ResetTimer()

			hw := HashWriter128{}
			for i := 0; i < b.N; i++ {
				var result [16]byte

				hw.Reset()
				hw.AddBytes(buf)
				hw.Sum128().Write(result[:])
			}
		})
	}
}

func BenchmarkHash128Branches(b *testing.B) {
	for length := 0; length <= 16; length++ {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf := make([]byte, length)
			b.SetBytes(int64(length))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				Hash128(buf)
			}
		})
	}
}

func Benchmark128Sizes(b *testing.B) {
	buf := make([]byte, 8192)
	for length := 32; length <= cap(buf); length *= 2 {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf = buf[:length]
			b.SetBytes(int64(length))
			b.ResetTimer()

			hw := HashWriter128{}
			for i := 0; i < b.N; i++ {
				var result [16]byte

				hw.Reset()
				hw.AddBytes(buf)
				hw.Sum128().Write(result[:])
			}
		})
	}
}

func BenchmarkHash128Sizes(b *testing.B) {
	buf := make([]byte, 8192)
	for length := 32; length <= cap(buf); length *= 2 {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			buf = buf[:length]
			b.SetBytes(int64(length))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				Hash128(buf)
			}
		})
	}
}
