package telegock

import (
	"io"
	"strings"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type Suite struct {
	suite.Suite
}

// NoPending checks if gock has pending requests every 10ms for maximum of given timeout.
// Default timeout is 2000ms.
func (suite *Suite) NoPending(timeout ...time.Duration) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	if len(timeout) == 0 {
		timeout = []time.Duration{2000 * time.Millisecond}
	}
	timeoutCh := time.After(timeout[0])

	for {
		select {
		case <-timeoutCh:
			pend := gock.Pending()
			if len(pend) == 0 {
				return
			}
			methods := make([]string, len(pend))
			for i, p := range pend {
				methods[i] = strings.TrimPrefix(p.Request().URLStruct.String(), base)
			}
			suite.Failf("Pending requests found", "Penging requests: %v", methods)
			return
		case <-ticker.C:
			if len(gock.Pending()) == 0 {
				return
			}
		}
	}
}

// NoUnmatched checks if gock has unmatched requests
func (suite *Suite) NoUnmatched() {
	time.Sleep(500 * time.Millisecond)
	unmatched := gock.GetUnmatchedRequests()
	if len(unmatched) == 0 {
		return
	}
	var methods []string
	for _, u := range unmatched {
		// skip for getUpdates
		if strings.HasSuffix(u.URL.String(), "/getUpdates") {
			continue
		}
		parts := strings.Split(u.URL.Path, "/")
		methods = append(methods, parts[len(parts)-1])
	}
	if len(methods) == 0 {
		return
	}
	suite.Failf("Unmatched requests found", "Unmatched requests: %v", methods)
}

func (suite *Suite) Decode(r io.ReadCloser) gjson.Result {
	b, err := io.ReadAll(r)
	suite.Require().NoError(err)
	suite.Require().NoError(r.Close())
	suite.Require().True(gjson.ValidBytes(b))
	return gjson.ParseBytes(b)
}
