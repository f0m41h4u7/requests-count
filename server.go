package main

import (
	"bufio"
	"fmt"
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
	// Buffer to save last stats
	buffer []int
	// Iterator over the buffer
	iter int
	// Current value
	currentCount int
}

type Server struct {
	http *http.Server

	statsBuffSize int
	statsFilename string
	mx            sync.Mutex
	stats         Stats
}

func NewServer(host, port, statsFile string, statsBuffSize int) *Server {
	serv := &Server{
		http: &http.Server{
			Addr: net.JoinHostPort(host, port),
		},
		statsBuffSize: statsBuffSize,
		statsFilename: statsFile,
		stats: Stats{
			buffer:       make([]int, statsBuffSize),
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
		s.stats.buffer[s.stats.iter]++
		/*fmt.Printf("statsIncrement=%d, iter=%d; [", s.stats.buffer[s.stats.iter], s.stats.iter)
		for _, b := range s.stats.buffer {
			fmt.Printf("%d, ", b)
		}
		fmt.Printf("]\n")*/
		s.mx.Unlock()
	}
}

// Periodically called method to update data in stats buffer
func (s *Server) UpdateStats() {
	s.mx.Lock()
	s.stats.currentCount += s.stats.buffer[s.stats.iter]
	s.stats.iter = ((s.stats.iter+1)%s.statsBuffSize + s.statsBuffSize) % s.statsBuffSize
	s.stats.buffer[s.stats.iter] = 0 - s.stats.buffer[s.stats.iter]
	//	fmt.Printf("UPDATE STATS: currentCount=%d, nextIdx=%d\n", s.stats.currentCount, s.stats.iter)
	s.mx.Unlock()
}

// Run starts Server
func (s *Server) Run() error {
	file, err := os.OpenFile(s.statsFilename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	num := 0
	i := 0
	for scanner.Scan() {
		num, err = strconv.Atoi(scanner.Text())
		if err != nil {
			return err
		}
		if i == s.statsBuffSize {
			break
		}
		s.stats.buffer[i] = num
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return s.http.ListenAndServe()
}

func (s *Server) Stop() error {
	var str string
	s.mx.Lock()
	defer s.mx.Unlock()
	for _, num := range s.stats.buffer {
		str += fmt.Sprintf("%d\n", num)
	}
	err := os.WriteFile(s.statsFilename, []byte(str), 0644)
	if err != nil {
		log.Printf("failed to write stats to file: %s", err.Error())
	}
	return s.http.Close()
}

func (s *Server) helloworld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello world!"})
}

func (s *Server) getStats(c *gin.Context) {
	s.mx.Lock()
	defer s.mx.Unlock()
	c.JSON(http.StatusOK, gin.H{"interval": (time.Duration(statsInterval) * time.Second).String(), "get requests number": strconv.Itoa(s.stats.currentCount)})
}
