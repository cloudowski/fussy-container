package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	version = "0.1.0"
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
		"invincible": 1,
	}
	imReady bool = false
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
	d := initdelay[delayType]
	time.Sleep(time.Duration(d) * time.Second)
	imReady = true
	fmt.Printf("Ready after delay type '%s' of %.2fs\n", delayType, d)
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
			if r < chance {
				panic(fmt.Sprintf("Killed! (drawn=%f, chance=%s, probability=%f)\n", r, crashChanceType, chance))
			} else {
				fmt.Printf("Missed - lucky you :-) (drawn=%f, chance=%s, probability=%f)\n", r, crashChanceType, chance)
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

func main() {
	conMeta.newMeta()
	conMeta.getIp()
	counter = 0
	delay := flag.String("delay", "bolt", fmt.Sprintf("delay type: zero, bolt, veyron, lovejava"))
	crash := flag.String("crash", "lotto", fmt.Sprintf("crash type: lotto, coin, badluck, invincible"))
	flag.Parse()
	go getReady(*delay)
	go crashMe(*crash)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", handler)
	// http.HandleFunc("/json", jsonHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
