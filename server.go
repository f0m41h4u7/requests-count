package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Stats struct {
	buffer       []int
	filename     string
	iter         int
	currentCount int
}

type Server struct {
	http *http.Server

	statsBuffSize int
	mx            sync.Mutex
	stats         Stats
}

func NewServer(host, port, statsFile string, statsBuffSize int) *Server {
	serv := &Server{
		http: &http.Server{
			Addr: net.JoinHostPort(host, port),
		},
		statsBuffSize: statsBuffSize,
		stats: Stats{
			buffer:       make([]int, statsBuffSize),
			filename:     statsFile,
			iter:         0,
			currentCount: 0,
		},
	}
	serv.http.Handler = serv.setupRouter()

	return serv
}

func (s *Server) setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(s.statsIncrement())
	router.GET("/helloworld", s.helloworld)

	return router
}

func (s *Server) statsIncrement() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			return
		}
		s.mx.Lock()
		s.stats.buffer[s.stats.iter]++
		fmt.Printf("statsIncrement=%d, iter=%d; [", s.stats.buffer[s.stats.iter], s.stats.iter)
		for _, b := range s.stats.buffer {
			fmt.Printf("%d, ", b)
		}
		fmt.Printf("]\n")
		s.mx.Unlock()
	}
}

func (s *Server) UpdateStats() {
	s.mx.Lock()
	s.stats.currentCount += s.stats.buffer[s.stats.iter]
	s.stats.iter = ((s.stats.iter+1)%s.statsBuffSize + s.statsBuffSize) % s.statsBuffSize
	s.stats.buffer[s.stats.iter] = 0 - s.stats.buffer[s.stats.iter]
	fmt.Printf("UPDATE STATS TICKER: currentCount=%d, nextIdx=%d\n", s.stats.currentCount, s.stats.iter)
	s.mx.Unlock()
}

func (s *Server) Run() error {
	return s.http.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.http.Close()
}

func (s *Server) helloworld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello world!"})
}
