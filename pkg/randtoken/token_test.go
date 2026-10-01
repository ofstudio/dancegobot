package randtoken

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestAlphabet(t *testing.T) {
	require.Equal(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", alphabet)
}

func TestNew(t *testing.T) {
	for _, n := range []int{0, 1, 2, 4, 8, 9, 10, 11, 12, 16, 64, 65, 256} {
		t.Run(fmt.Sprintf("length=%d", n), func(t *testing.T) {
			token := New(n)
			require.Len(t, token, n)
			require.Equal(t, n, utf8.RuneCountInString(token))
			require.True(t, tokenInAlphabet(token))
		})
	}
}

func TestNewNegativeLength(t *testing.T) {
	require.Panics(t, func() { New(-1) })
}

func TestNewDistinct(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 32; i++ {
		token := New(64)
		require.False(t, seen[token], "duplicate token")
		seen[token] = true
	}
}

func TestRandomInt63(t *testing.T) {
	for i := 0; i < 32; i++ {
		require.GreaterOrEqual(t, randomInt63(), int64(0))
	}
}

func TestNewConcurrent(t *testing.T) {
	const (
		goroutines = 32
		iterations = 500
		tokenLen   = 16
	)

	var wg sync.WaitGroup
	errs := make(chan error, goroutines*iterations)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				token := New(tokenLen)
				if len(token) != tokenLen {
					errs <- fmt.Errorf("token length = %d, want %d", len(token), tokenLen)
					return
				}
				if !tokenInAlphabet(token) {
					errs <- fmt.Errorf("token %q contains symbols outside alphabet", token)
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
}

func tokenInAlphabet(token string) bool {
	for _, r := range token {
		if !strings.ContainsRune(alphabet, r) {
			return false
		}
	}
	return true
}
