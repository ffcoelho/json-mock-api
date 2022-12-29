package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ffcoelho/jma/keys"
)

var tty *keys.TTY
var codes []int = []int{200, 201, 202, 204, 301, 302, 304, 400, 401, 403, 404, 409, 410, 500, 501, 503}

var startServer bool
var lastPrint int = -1
var statusCodeIdx int

var port uint16 = 9000
var delay int
var statusCode int = http.StatusOK

func init() {
	startServer = processArgs()
	if startServer {
		setupExitHandler()
		setupKeyboardHandler()
	}
}

func main() {
	if startServer {
		printHeader()
		http.HandleFunc("/", handler)
		ip := getIP()
		printInfo(ip.String())
		addr := fmt.Sprintf(":%d", port)
		http.ListenAndServe(addr, nil)
	} else {
		printHelp()
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	lastPrint = 0
	now := time.Now()
	fmt.Printf("\n%s %d %s %s", now.Format("15:04:05.000"), statusCode, req.Method, req.URL.Path)
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	time.Sleep(time.Duration(delay) * time.Millisecond)
	w.Write([]byte("Ok"))
}

func toggleDelay() {
	br := ""
	if lastPrint != 1 {
		lastPrint = 1
		br = "\n"
	}
	fmt.Printf("%s\r                  ", br)
	switch delay {
	case 0:
		delay = 600
	case 600:
		delay = 1200
	case 1200:
		delay = 2400
	default:
		delay = 0
		fmt.Printf("\rDELAY: OFF")
		return
	}
	fmt.Printf("\rDELAY: %dms", delay)
}

func toggleStatus(i int) {
	br := ""
	if lastPrint != 2 {
		lastPrint = 2
		br = "\n"
	}

	statusCodeIdx += i
	if statusCodeIdx == -1 {
		statusCodeIdx = len(codes) - 1
	} else if statusCodeIdx == len(codes) {
		statusCodeIdx = 0
	}
	statusCode = codes[statusCodeIdx]
	fmt.Printf("%s\rSTATUS CODE: %d", br, codes[statusCodeIdx])
}

func printHeader() {
	fmt.Printf("JSON Mock API v1.0\n\n")
	fmt.Printf(" Key   Command               Default\n")
	fmt.Printf(" a, s  change status code    200\n")
	fmt.Printf(" d     toggle delay          OFF\n")
}

func printInfo(ip string) {
	fmt.Printf("\nListening on http://localhost:%d\n", port)
	fmt.Printf("             http://%s:%d\n", ip, port)
}

func printHelp() {
	fmt.Println("Usage: ./mock <port>")
	fmt.Println("Example: ./mock 3000    (default: 9000)")
}

func setupExitHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nJSON Mock API is down.")
		tty.Close()
		os.Exit(0)
	}()
}

func setupKeyboardHandler() {
	go func() {
		tty = keys.Open()

		for {
			key := tty.ReadKey()
			switch key {
			case "a":
				toggleStatus(-1)
			case "d":
				toggleDelay()
			case "s":
				toggleStatus(1)
			}
		}
	}()
}

func processArgs() bool {
	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
			return false
		} else {
			pUi64, err := strconv.ParseUint(args[0], 10, 64)
			if err == nil {
				port = uint16(pUi64)
			}
		}
	}
	return true
}

func getIP() net.IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	ips := make([]net.IP, 0)
	for _, i := range ifaces {
		addrs, e := i.Addrs()
		if e != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			ip = ip.To4()
			if ip == nil || ip.String() == "127.0.0.1" {
				continue
			}
			ips = append(ips, ip)
		}
	}
	return ips[0]
}
