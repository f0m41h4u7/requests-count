package main

import (
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestServer(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		statsQuant := 1
		statsQty := 60

		var mutex sync.Mutex
		srv := &Server{
			http: &http.Server{
				Addr: "127.0.0.1:1337",
			},
			statsBuffSize: statsQty,
			stats: Stats{
				buffer:       make([]int, statsQty),
				iter:         0,
				currentCount: 0,
			},
		}

		srv.http.Handler = srv.setupRouter()

		done := make(chan struct{}, 1)
		requestCnt := 0

		ticker := time.NewTicker(time.Duration(statsQuant) * time.Second)
		timer := time.NewTimer(time.Duration(statsQuant*statsQty) * time.Second)
		stop := false

		go func(done <-chan struct{}, srv *Server, cnt *int) {
			for {
				select {
				case <-done:
					return
				default:
					mutex.Lock()
					w := performRequest(srv.http.Handler, "GET", "/helloworld", nil)
					mutex.Unlock()
					require.Equal(t, http.StatusOK, w.Code)
					requestCnt++
					time.Sleep(time.Duration(rand.Intn(4)-1) * time.Second)
				}
			}
		}(done, srv, &requestCnt)

		for !stop {
			select {
			case <-ticker.C:
				mutex.Lock()
				srv.UpdateStats()
				mutex.Unlock()
			case <-timer.C:
				close(done)
				stop = true
			}
		}
		time.Sleep(time.Duration(5) * time.Second)

		require.Less(t, abs(requestCnt-srv.stats.currentCount), 5, "Wrong number of requests was counted. Expected=%d, actual=%d", requestCnt, srv.stats.currentCount)
	})
}
