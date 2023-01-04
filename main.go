package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var showHelp bool
var port uint16
var prefix string
var localIP net.IP

var lastPrint int = -1

func init() {
	help := flag.Bool("help", false, "Help")
	portFlag := flag.Int("port", 9000, "Server port")
	prefixFlag := flag.String("prefix", "", "Routes prefix")

	flag.Parse()
	port = uint16(*portFlag)
	prefix = processPrefixFlag(*prefixFlag)

	if *help || (len(os.Args) > 1 && os.Args[1] == "help") {
		showHelp = true
	}
}

func main() {
	if showHelp {
		printHelp()
		return
	}

	getIP()
	readMockFile()
	setupExitHandler()
	setupKeyboardHandler()
	printHeader()
	http.HandleFunc("/", handler)
	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, prefix) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	reqPath := processRequestPath(req.URL.Path)
	pathSep := ""
	if !strings.HasPrefix(reqPath, "/") {
		pathSep = "/"
	}
	reqEls := processRouteElements(reqPath)
	response, resStatus := processResponse(reqEls, req.Method)
	lastPrint = 0
	now := time.Now()
	w.Header().Set("Content-Type", "application/json")
	fmt.Printf("\n%s %d %s %s%s %s", now.Format("15:04:05.000"), resStatus, req.Method, pathSep, reqPath, req.RemoteAddr)
	time.Sleep(time.Duration(delay) * time.Millisecond)
	if response == nil {
		msg := fmt.Sprintf("Json Mock API: \"%s %s\" not found", req.Method, reqPath)
		http.Error(w, msg, http.StatusNotFound)
		return
	}
	w.WriteHeader(resStatus)
	json.NewEncoder(w).Encode(response)
}

func processPrefixFlag(p string) string {
	els := strings.Split(p, "/")
	if len(els) == 0 {
		return "/"
	}
	cleanEls := ""
	for _, el := range els {
		if len(el) > 0 {
			cleanEls = cleanEls + "/" + el
		}
	}
	if len(cleanEls) == 0 {
		return "/"
	}
	return cleanEls
}

func processRequestPath(path string) string {
	p := strings.SplitAfterN(path, prefix, 2)[1]
	if len(p) == 0 || p == "/" {
		p = "/"
	} else {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

func getIP() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal("ERROR: net\n")
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
			if ip == nil || !strings.HasPrefix(ip.String(), "192.") {
				continue
			}
			ips = append(ips, ip)
		}
	}
	if len(ips) == 0 {
		return
	}
	localIP = ips[0]
}

func printHeader() {
	fmt.Printf("JSON Mock API v1.0\n\n")
	fmt.Printf("  a, s    change status code\n")
	fmt.Printf("  d       toggle delay\n")
	fmt.Printf("  ctrl+c  stop server\n\n")
	fmt.Printf("For more info, run help.\n\n")
	for _, r := range routes {
		methods := ""
		for _, m := range r.methods {
			methods = methods + " " + m.method
		}
		es := strings.Repeat(" ", apiPathsWidth-len(r.path))
		fmt.Printf("%s%s  %s\n", r.path, es, methods)
	}
	fmt.Printf("\nListening on http://localhost:%d%s\n", port, prefix)
	if localIP != nil {
		fmt.Printf("             http://%s:%d%s\n", localIP.String(), port, prefix)
	}
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
