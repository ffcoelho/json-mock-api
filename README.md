# JSON Mock API
An easy way to run a mock API.
* JSON template
* Control HTTP status and delay
* Record of incoming requests
  
*This application was made for educational and development purposes.*
## Download
* [Linux](https://github.com/ffcoelho/json-mock-api/tree/main/zip/linux.zip) - x64
* [Windows](https://github.com/ffcoelho/json-mock-api/tree/main/zip/windows.zip) - x64
  
## Usage
```
$ ./mock [-port] [-prefix]

> mock.exe [-port] [-prefix]  (windows)


OPTIONS
  -port, --port          Override default port (9000)
  -prefix, --prefix      Add path prefix to all routes
  help, -help, --help    Show help

EXAMPLES
  $ ./mock
  $ ./mock --port=3000
  $ ./mock -port 3000 -prefix api/v1

SHORTCUTS
  a, s      Change status code
  d         Toggle delay
  ctrl+c    Stop server
```
## mock.json
```
{
  "path": {
    "method": {
      "code": RESPONSE
    }
  }
}

EXAMPLE
{
  "/": {
    "GET": {
      "200": { "mock": true }
    }
  },
  "/books": {
    "GET": {
      "200": { "books": [] }
    }
  },
  "/books/:id/reviews": {
    "POST": {
      "201": { "error": false },
      "400": { "error": true }
    },
    "GET": {
      "200": { "reviews": [] }
    }
  }
}
```
