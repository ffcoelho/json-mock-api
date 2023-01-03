package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type MockRoute struct {
	path    string
	pathEls []string
	methods []MockMethod
}

type MockMethod struct {
	method    string
	responses []MockResponse
}

type MockResponse struct {
	code    string
	payload any
}

var routes []MockRoute
var methods []string = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
var codes []int = []int{1}
var successCodes []int = []int{200, 201, 202, 204}
var apiPathsWidth int

func processRouteElements(route string) []string {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
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

func processResponse(els []string, m string) (interface{}, int) {
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
					successCode := statusCode == 1
					if fmt.Sprint(statusCode) == c.code || (successCode && isSuccessCode(c.code)) {
						rc, _ := strconv.Atoi(c.code)
						return c.payload, rc
					}
				}
				resCode := statusCode
				if resCode == 1 {
					resCode = 200
				}
				return map[string]interface{}{}, resCode
			}
		}
		break
	}
	return nil, 404
}

func readMockFile() error {
	mockFile, err := os.Open("mock.json")
	checkError(err, "open_file")
	defer mockFile.Close()

	byteValue, err := io.ReadAll(mockFile)
	checkError(err, "read_file")
	var result map[string]map[string]map[string]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)
	checkError(err, "read_file")

	for route, methods := range result {
		routeElements := processRouteElements(route)
		if len(routeElements) == 0 {
			continue
		}
		var mockRoute MockRoute
		mockRoute.path = route
		mockRoute.pathEls = routeElements
		for method, responses := range methods {
			if invalidMethod(method) {
				continue
			}
			var mockMethod MockMethod
			mockMethod.method = method
			for code, payload := range responses {
				codeInt, err := validateStatusCode(code)
				if err != nil {
					continue
				}
				var mockResponse MockResponse
				mockResponse.code = code
				mockResponse.payload = payload
				mockMethod.responses = append(mockMethod.responses, mockResponse)
				if !containsCode(codeInt) {
					codes = append(codes, codeInt)
				}
			}
			sort.Slice(mockMethod.responses,
				func(i, j int) bool {
					return mockMethod.responses[j].code > mockMethod.responses[i].code
				})
			mockRoute.methods = append(mockRoute.methods, mockMethod)
		}
		routes = append(routes, mockRoute)
		if len(mockRoute.path) > apiPathsWidth {
			apiPathsWidth = len(mockRoute.path)
		}
	}
	for ri, route := range routes {
		sort.Slice(route.methods,
			func(i, j int) bool {
				return routes[ri].methods[j].method > routes[ri].methods[i].method
			})
	}
	sort.Slice(routes,
		func(i, j int) bool {
			return routes[j].path > routes[i].path
		})
	return nil
}

func invalidMethod(key string) bool {
	for _, method := range methods {
		if method == key {
			return false
		}
	}
	return true
}

func validateStatusCode(c string) (int, error) {
	if len(c) != 3 {
		return 0, fmt.Errorf("len")
	}
	ns, err := strconv.Atoi(c)
	if err != nil {
		return 0, fmt.Errorf("int")
	}
	if ns < 100 || ns > 599 {
		return ns, nil
	}
	return ns, nil
}

func containsCode(c int) bool {
	for _, code := range codes {
		if c == code {
			return true
		}
	}
	return false
}

func isSuccessCode(c string) bool {
	for _, code := range successCodes {
		if c == fmt.Sprint(code) {
			return true
		}
	}
	return false
}

func checkError(e error, s string) {
	if e != nil {
		switch s {
		case "open_file":
			log.Fatal("ERROR: mock.json not found. Run help for more info.\n")
		case "read_file":
			log.Fatal("ERROR: invalid mock.json. Run help for more info.\n")
		default:
			log.Fatal("ERROR: something went wrong.\n")
		}
	}
}
