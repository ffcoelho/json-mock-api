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
var codes []int = []int{200, 201, 204, 301, 302, 304, 400, 401, 403, 404, 409, 410, 500, 501}

var lastPrint int = -1
var mockDelay int
var mockStatus int = http.StatusOK
var mockStatusIdx int
var port uint16 = 9000

func init() {
	setupExitHandler()
	setupKeyboardHandler()
}

func main() {
	stop := processArgs()
	if stop == false {
		printHeader()
		http.HandleFunc("/", handler)
		ip := getIP()
		printInfo(ip.String())
		addr := fmt.Sprintf(":%d", port)
		http.ListenAndServe(addr, nil)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	if lastPrint != 0 {
		lastPrint = 0
	}

	now := time.Now()
	fmt.Printf("\n%s %d %s %s", now.Format("15:04:05.000"), mockStatus, req.Method, req.URL.Path)
	w.WriteHeader(mockStatus)
	w.Header().Set("Content-Type", "application/json")
	time.Sleep(time.Duration(mockDelay) * time.Millisecond)
	w.Write([]byte("Ok"))
}

func toggleDelay() {
	br := ""
	if lastPrint != 1 {
		lastPrint = 1
		br = "\n"
	}
	fmt.Printf("%s\r                  ", br)
	switch mockDelay {
	case 0:
		mockDelay = 600
	case 600:
		mockDelay = 1200
	case 1200:
		mockDelay = 2400
	default:
		mockDelay = 0
		fmt.Printf("\rDELAY: OFF")
		return
	}
	fmt.Printf("\rDELAY: %dms", mockDelay)
}

func toggleStatus(i int) {
	br := ""
	if lastPrint != 2 {
		lastPrint = 2
		br = "\n"
	}

	mockStatusIdx += i
	if mockStatusIdx == -1 {
		mockStatusIdx = len(codes) - 1
	}
	if mockStatusIdx == len(codes) {
		mockStatusIdx = 0
	}
	mockStatus = codes[mockStatusIdx]
	fmt.Printf("%s\rSTATUS CODE: %d", br, codes[mockStatusIdx])
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
			printHelp()
			return true
		} else {
			pUi64, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return false
			}
			port = uint16(pUi64)
		}
	}
	return false
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
