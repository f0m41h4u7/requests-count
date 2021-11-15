package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Stats struct {
	// Size of Buffer
	BuffSize int `json:"buffSize"`
	// Buffer to save last stats
	Buffer []int `json:"Buffer"`
	// Iterator over the Buffer
	Iter int `json:"Iter"`
	// Current value
	CurrentCount int `json:"CurrentCount"`
}

type Server struct {
	http *http.Server

	statsFilename string
	mx            sync.Mutex
	stats         Stats
}

func NewServer(host, port, statsFile string, statsBuffSize int) *Server {
	serv := &Server{
		http: &http.Server{
			Addr: net.JoinHostPort(host, port),
		},
		statsFilename: statsFile,
		stats: Stats{
			BuffSize:     statsBuffSize,
			Buffer:       make([]int, statsBuffSize),
			Iter:         0,
			CurrentCount: 0,
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
	router.GET("/stats", s.getStats)

	return router
}

// Middleware handler for counting GET requests
func (s *Server) statsIncrement() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			return
		}
		s.mx.Lock()
		s.stats.Buffer[s.stats.Iter]++
		/*fmt.Printf("statsIncrement=%d, Iter=%d; [", s.stats.Buffer[s.stats.Iter], s.stats.Iter)
		for _, b := range s.stats.Buffer {
			fmt.Printf("%d, ", b)
		}
		fmt.Printf("]\n")*/
		s.mx.Unlock()
	}
}

// Periodically called method to update data in stats Buffer
func (s *Server) UpdateStats() {
	s.mx.Lock()
	s.stats.CurrentCount += s.stats.Buffer[s.stats.Iter]
	s.stats.Iter = ((s.stats.Iter+1)%s.stats.BuffSize + s.stats.BuffSize) % s.stats.BuffSize
	s.stats.Buffer[s.stats.Iter] = 0 - s.stats.Buffer[s.stats.Iter]
	//	fmt.Printf("UPDATE STATS: CurrentCount=%d, nextIdx=%d\n", s.stats.CurrentCount, s.stats.Iter)
	s.mx.Unlock()
}

// Run starts Server
func (s *Server) Run() error {
	file, err := ioutil.ReadFile(s.statsFilename)
	if err != nil {
		log.Printf("failed to read stats from file: %s", err.Error())
	} else {
		err = json.Unmarshal([]byte(file), &s.stats)
		if err != nil {
			log.Fatalf("failed to parse stats: %s", err.Error())
		}
	}

	return s.http.ListenAndServe()
}

func (s *Server) Stop() error {
	s.mx.Lock()
	defer s.mx.Unlock()
	dataBytes, err := json.Marshal(s.stats)
	if err != nil {
		log.Printf("failed to marshal stats: %s", err.Error())
	} else {
		err = os.WriteFile(s.statsFilename, dataBytes, 0644)
		if err != nil {
			log.Printf("failed to write stats to file: %s", err.Error())
		}
	}
	return s.http.Close()
}

func (s *Server) helloworld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello world!"})
}

func (s *Server) getStats(c *gin.Context) {
	s.mx.Lock()
	defer s.mx.Unlock()
	c.JSON(http.StatusOK, gin.H{"interval": (time.Duration(statsInterval) * time.Second).String(), "get requests number": strconv.Itoa(s.stats.CurrentCount)})
}
