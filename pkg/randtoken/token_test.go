package randtoken

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

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
