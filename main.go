package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	version = "0.5"
)

var (
	conMeta   meta
	counter   int
	initdelay = map[string]float32{
		"zero":     0.0,
		"bolt":     9.58,
		"veyron":   16.7,
		"lovejava": 50.0,
	}
	crashChance = map[string]float32{
		"lotto":      float32(1.0 / 49.0),
		"coin":       0.5,
		"badluck":    0.9,
		"invincible": 0,
	}
	// reverse chance - if true container will survive instead of crash with the defined probability
	reverseChance bool = false
	imReady       bool = false
)

type meta struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip",omitempty`
	ExtIP    string `json:"extip",omitempty`
}

func (m *meta) newMeta() {
	m.Hostname, _ = os.Hostname()
}

func (m *meta) getIp() {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error getting interfaces: %s", err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Printf("Error getting ip from %v: %s", i, err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// fmt.Printf("%s\n", ip)
			if !ip.IsLoopback() {
				m.IP = fmt.Sprintf("%s", ip)
				break
			}
		}
	}

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if imReady {
		w.WriteHeader(200)
		fmt.Fprintf(w, "ok")
	} else {
		w.WriteHeader(503)
		fmt.Fprintf(w, "not ready")
	}
}

func getReady(delayType string) {
	if d, ok := initdelay[delayType]; ok {
		fmt.Printf("Preparing (delay type='%s', duration=%.2fs)\n", delayType, d)
		time.Sleep(time.Duration(d) * time.Second)
		imReady = true
		fmt.Printf("Ready!\n")
	} else {
		panic(fmt.Sprintf("Invalid delay type: %s", delayType))
	}
}

func crashMe(crashChanceType string) {
	chance := crashChance[crashChanceType]
	// i := 0
	for {
		if !imReady {
			time.Sleep(time.Second)
		} else {
			// shoot!
			src := rand.NewSource(time.Now().UnixNano())
			rnd := rand.New(src)
			r := rnd.Float32()
			c := chance
			if reverseChance {
				c = 1 - chance
			}
			if r < c {
				panic(fmt.Sprintf("Killed! (drawn=%f, chance=%s, probability=%f)\n", r, crashChanceType, c))
			} else {
				fmt.Printf("Missed - lucky you :-) (drawn=%f, chance=%s, probability=%f)\n", r, crashChanceType, c)
				return
				// i += 1
				// time.Sleep(20 * time.Millisecond)
				// fmt.Printf("[%d] ", i)
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	counter += 1
	fmt.Fprintf(w, "%-15s %s\n", "Hostname:", conMeta.Hostname)
	fmt.Fprintf(w, "%-15s %s\n", "IP:", conMeta.IP)
	fmt.Fprintf(w, "%-15s %s\n", "External IP:", conMeta.ExtIP)
	fmt.Fprintf(w, "%-15s %d\n", "Request count:", counter)
	fmt.Fprintf(w, "%-15s %s\n", "Version:", version)
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	counter += 1
	cj, _ := json.Marshal(conMeta)
	fmt.Fprintf(w, "%s", string(cj))
}

func configFromEnv(key string) (string, error) {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if strings.ToLower(pair[0]) == key {
			return pair[1], nil
		}
	}
	return "", errors.New(fmt.Sprintf("Key %s not found in environment variables", key))
}

func sigHandler(sigs chan os.Signal, done chan bool) {
	sig := <-sigs
	fmt.Printf("Received and handling signal %s\n", sig)
	done <- true
	os.Exit(0)
}

func main() {
	conMeta.newMeta()
	conMeta.getIp()
	counter = 0

	delay, err1 := configFromEnv("delay")
	if err1 != nil {
		flag.StringVar(&delay, "delay", "bolt", fmt.Sprintf("delay type: zero, bolt, veyron, lovejava"))
	}
	crash, err2 := configFromEnv("crash")
	if err2 != nil {
		flag.StringVar(&crash, "crash", "lotto", fmt.Sprintf("crash type: lotto, coin, badluck, invincible"))
	}

	flag.BoolVar(&reverseChance, "reverse", reverseChance, fmt.Sprintf("reverse action - survive instead of crash with defined probability"))

	flag.Parse()
	go getReady(delay)
	go crashMe(crash)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go sigHandler(sigs, done)

	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/", handler)
	// http.HandleFunc("/json", jsonHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
