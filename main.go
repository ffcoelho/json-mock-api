package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ffcoelho/jma/keys"
)

type Route struct {
	path    string
	pathEls []string
	methods []Method
}

type Method struct {
	method    string
	responses []MockResponse
}

type MockResponse struct {
	code    string
	payload any
}

var tty *keys.TTY
var methods []string = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD", "CONNECT", "TRACE"}
var codes []int = []int{200, 201, 202, 204, 301, 302, 304, 400, 401, 403, 404, 409, 410, 500, 501, 503}

var showHelp bool
var lastPrint int = -1
var statusCodeIdx int

var port uint16
var prefix string
var delay int
var statusCode int = http.StatusOK
var routes []Route
var apiPaths []string

func init() {
	help := flag.Bool("help", false, "Help")
	serverPort := flag.Int("port", 9000, "Server port")
	apiPrefix := flag.String("prefix", "", "Routes prefix")
	flag.Parse()
	port = uint16(*serverPort)
	prefix = *apiPrefix

	if *help || (len(os.Args) > 1 && os.Args[1] == "help") {
		showHelp = true
	}
}

func main() {
	if showHelp {
		printHelp()
		return
	}

	ip := getIP()
	readMockFile()
	setupExitHandler()
	setupKeyboardHandler()
	http.HandleFunc("/", handler)
	printInfo(ip.String())
	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, "/"+prefix) {
		http.Error(w, "Error: invalid prefix", http.StatusNotFound)
		return
	}

	reqPath := strings.SplitAfter(req.URL.Path, prefix)[1]
	reqPath = strings.TrimSuffix(reqPath, "/")
	if len(reqPath) == 0 {
		reqPath = "/"
	}
	reqEls := processRouteElements(reqPath)
	response := processResponse(reqEls, req.Method)
	lastPrint = 0
	now := time.Now()
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	fmt.Printf("\n%s %d %s %s %s", now.Format("15:04:05.000"), statusCode, req.Method, reqPath, req.RemoteAddr)
	time.Sleep(time.Duration(delay) * time.Millisecond)
	json.NewEncoder(w).Encode(response)
}

func processResponse(els []string, m string) interface{} {
	var response interface{}
routes:
	for _, route := range routes {
		if len(els) != len(route.pathEls) {
			continue
		}
		for i, pe := range route.pathEls {
			if els[i] != pe && pe != ":id" {
				continue routes
			}
		}
		for _, rm := range route.methods {
			if m == rm.method {
				for _, c := range rm.responses {
					if fmt.Sprint(statusCode) == c.code {
						response = c.payload
					}
				}
			}
		}
		break
	}
	if response != nil {
		return response
	}
	return map[string]interface{}{}
}

func toggleDelay() {
	br := ""
	if lastPrint != 1 {
		lastPrint = 1
		br = "\n"
	}
	fmt.Printf("%s\r              ", br)
	switch delay {
	case 0:
		delay = 600
	case 600:
		delay = 1200
	case 1200:
		delay = 2400
	default:
		delay = 0
		fmt.Printf("\rDELAY: off")
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

func setupExitHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
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
			case "s":
				toggleStatus(1)
			case "d":
				toggleDelay()
			}
		}
	}()
}

func getIP() net.IP {
	ifaces, err := net.Interfaces()
	checkSetupError(err, "net")

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

func checkSetupError(e error, s string) {
	if e != nil {
		switch s {
		case "open_file":
			fmt.Printf("ERROR: mock.json not found. Run help for more info.\n")
		case "read_file":
			fmt.Printf("ERROR: invalid mock.json. Run help for more info.\n")
		default:
			fmt.Printf("ERROR: something went wrong.\n")
		}
		os.Exit(0)
	}
}

func readMockFile() error {
	mockFile, err := os.Open("mock.json")
	checkSetupError(err, "open_file")
	defer mockFile.Close()

	byteValue, err := io.ReadAll(mockFile)
	checkSetupError(err, "read_file")
	var result map[string]map[string]map[string]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)
	checkSetupError(err, "read_file")

	for route, methods := range result {
		routeElements := processRouteElements(route)
		if len(routeElements) == 0 {
			continue
		}
		var mockRoute Route
		mockRoute.path = route
		mockRoute.pathEls = routeElements
		for method, responses := range methods {
			if invalidMethod(method) {
				continue
			}
			var mockMethod Method
			mockMethod.method = method
			for code, payload := range responses {
				if invalidStatusCode(code) {
					continue
				}
				var mockResponse MockResponse
				mockResponse.code = code
				mockResponse.payload = payload
				mockMethod.responses = append(mockMethod.responses, mockResponse)
			}
			mockRoute.methods = append(mockRoute.methods, mockMethod)
		}
		routes = append(routes, mockRoute)
		apiPaths = append(apiPaths, mockRoute.path)
	}
	sort.Strings(apiPaths)
	return nil
}

func processRouteElements(route string) []string {
	if !strings.HasPrefix(route, "/") {
		return []string{}
	}
	if route == "/" {
		return []string{"*root*"}
	}
	els := strings.Split(route, "/")[1:]
	for _, el := range els {
		if len(el) == 0 {
			return []string{}
		}
	}
	return els
}

func invalidMethod(key string) bool {
	for _, method := range methods {
		if method == key {
			return false
		}
	}
	return true
}

func invalidStatusCode(statusCode string) bool {
	sc, err := strconv.Atoi(statusCode)
	if err != nil {
		return true
	}
	for _, code := range codes {
		if code == sc {
			return false
		}
	}
	return true
}

func printInfo(ip string) {
	fmt.Printf("JSON Mock API v1.0\n\n")
	fmt.Printf("  a, s    change status code\n")
	fmt.Printf("  d       toggle delay\n")
	fmt.Printf("  ctrl+c  stop server\n\n")
	fmt.Printf("For more info, run help.\n\n")
	for _, path := range apiPaths {
		fmt.Printf("%s\n", path)
	}
	prefixInfo := ""
	if len(prefix) > 0 {
		prefixInfo = "/" + prefix
	}
	fmt.Printf("\nListening on http://localhost:%d%s\n", port, prefixInfo)
	fmt.Printf("             http://%s:%d%s\n", ip, port, prefixInfo)
}

func printHelp() {
	fmt.Printf("JSON Mock API v1.0 Help\n\n")
	fmt.Printf("USAGE\n\n")
	fmt.Printf("  $ ./mock [--port{9000}] [--prefix]\n\n")
	fmt.Printf("  - Examples:\n")
	fmt.Printf("    $ ./mock\n")
	fmt.Printf("    $ ./mock -port=3000 -prefix=api/v1\n\n")
	fmt.Printf("COMMANDS\n\n")
	fmt.Printf("  a, s  change status code\n")
	fmt.Printf("  d     toggle delay\n\n")
	fmt.Printf("MOCK ROUTES (mock.json)\n\n")
	fmt.Printf("  PATH: {\n")
	fmt.Printf("    METHOD: {\n")
	fmt.Printf("      CODE: PAYLOAD\n")
	fmt.Printf("    }\n")
	fmt.Printf("  }\n\n")
	fmt.Printf("  - Example:\n")
	fmt.Printf("    {\n")
	fmt.Printf("      \"/books\": {\n")
	fmt.Printf("        \"GET\": {\n")
	fmt.Printf("          \"200\": { \"books\": [] }\n")
	fmt.Printf("        }\n")
	fmt.Printf("      },\n")
	fmt.Printf("      \"/books/:id/reviews\": {\n")
	fmt.Printf("        \"POST\": {\n")
	fmt.Printf("          \"201\": { \"error\": false },\n")
	fmt.Printf("          \"400\": { \"error\": true }\n")
	fmt.Printf("        },\n")
	fmt.Printf("        \"GET\": {\n")
	fmt.Printf("          \"200\": { \"reviews\": [] }\n")
	fmt.Printf("        }\n")
	fmt.Printf("      }\n")
	fmt.Printf("    }\n\n")
}
