# JSON Mock API
An easy way to run a mock API.
* JSON template
* Control HTTP status and delay
* Record of incoming requests

## Download
* [Linux](https://github.com/ffcoelho/mock/zip/linux.zip)
* [Windows](https://github.com/ffcoelho/mock/zip/windows.zip)

## Usage
```
$ ./mock [-port] [-prefix]

OPTIONS
  -port, --port          Override default port (9000)
  -prefix, --prefix      Add path prefix to all routes
  help, -help, --help    Show help

EXAMPLES
  $ ./mock
  $ ./mock -port=3000 -prefix=api/v1

COMMANDS
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
      "200": []
    }
  }
}
```
