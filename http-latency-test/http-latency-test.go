package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/gosuri/uilive"
)

type Statistics struct {
	times   []time.Duration
	current time.Duration
	avg     time.Duration
	max     time.Duration
	min     time.Duration
	sd      time.Duration
	p50     time.Duration
	p75     time.Duration
	p90     time.Duration
	p95     time.Duration
	p99     time.Duration
}

func newStats() *Statistics {
	return &Statistics{
		times:   make([]time.Duration, 0),
		current: 0,
		avg:     0,
		max:     0,
		min:     math.MaxInt64,
		sd:      0,
		p50:     0,
		p75:     0,
		p90:     0,
		p95:     0,
		p99:     0,
	}
}

func (s *Statistics) addTime(t time.Duration) {
	s.current = t
	i, _ := slices.BinarySearch(s.times, t)
	s.times = slices.Insert(s.times, i, t)

	if t > s.max {
		s.max = t
	}
	if t < s.min {
		s.min = t
	}

	// Calculate avg and sd over seconds instead of nanoseconds
	// to avoid float64 overflow

	avgSeconds := 0.0
	for _, t := range s.times {
		avgSeconds += t.Seconds()
	}
	avgSeconds /= float64(len(s.times))
	s.avg = time.Duration(avgSeconds * float64(time.Second))

	sdSeconds := 0.0
	for _, t := range s.times {
		sdSeconds += (t.Seconds() - avgSeconds) * (t.Seconds() - avgSeconds)
	}
	sdSeconds = math.Sqrt(sdSeconds / float64(len(s.times)))
	s.sd = time.Duration(sdSeconds * float64(time.Second))

	s.p50 = s.times[len(s.times)/2]
	s.p75 = s.times[len(s.times)*3/4]
	s.p90 = s.times[len(s.times)*9/10]
	s.p95 = s.times[len(s.times)*19/20]
	s.p99 = s.times[len(s.times)*99/100]
}

func (s *Statistics) getString() string {
	return fmt.Sprintf(
		`Count: %d | Current: %dms
Min: %dms | Max: %dms | Avg: %dms +/- Std: %dms
p50: %dms | p75: %dms | p90: %dms | p95: %dms | p99: %dms
`,
		len(s.times), s.current.Milliseconds(), s.min.Milliseconds(), s.max.Milliseconds(),
		s.avg.Milliseconds(), s.sd.Milliseconds(), s.p50.Milliseconds(), s.p75.Milliseconds(),
		s.p90.Milliseconds(), s.p95.Milliseconds(), s.p99.Milliseconds(),
	)
}

func main() {
	url := flag.String("url", "", "URL to do a GET request to")
	sleepMillis := flag.Int("sleep", 500, "Time between request in milliseconds")
	count := flag.Int("count", 0, "Max number of requests")

	flag.Parse()

	if *url == "" {
		println("Error: missing url")
		flag.PrintDefaults()
		os.Exit(1)
	}

	stats := newStats()
	writer := uilive.New()
	defer writer.Stop()

	writer.Start()
	for {
		t, err := getRequest(*url)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}

		stats.addTime(t)
		fmt.Fprint(writer, stats.getString())

		if *count > 0 && len(stats.times) >= *count {
			break
		}

		time.Sleep(time.Duration(*sleepMillis) * time.Millisecond)
	}
}

func getRequest(url string) (time.Duration, error) {
	t0 := time.Now()
	_, err := http.Get(url)
	return time.Since(t0), err
}
