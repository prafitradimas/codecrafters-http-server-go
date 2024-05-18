package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type (
	Request struct {
		httpversion string
		host        string
		method      string
		path        string
		headers     map[string]string
		body        string
	}

	Response struct {
		httpversion string
		code        int
		status      string
		headers     []string
		body        string
	}
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	buf := make([]byte, 4096)
	conn.Read(buf)
	fmt.Println(string(buf))

	req, _ := parseRequest(string(buf))
	res := &Response{
		httpversion: "1.1",
		body:        req.body,
	}

	param, found := strings.CutPrefix(req.path, "/echo/")
	if req.path == "/" {
		res.code = 200
		write(conn, res)
	} else if req.path == "/user-agent" {
		res.code = 200
		res.body = req.headers["User-Agent"]
		res.headers = []string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %d", len(res.body))}
		write(conn, res)
	} else if found {
		res.code = 200
		res.body = param
		res.headers = []string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %d", len(res.body))}
		write(conn, res)
	} else {
		res.code = 404
		write(conn, res)
	}

}

func write(conn net.Conn, res *Response) {
	headers := ""
	if 0 < len(res.headers) {
		headers = strings.Join(res.headers, "\r\n")
	}

	status := httpstatus(res.code)
	respstr := fmt.Sprintf("HTTP/%s %d %s\r\n%s\r\n\r\n%s", res.httpversion, res.code, status, headers, res.body)
	_, err := conn.Write([]byte(respstr))

	if err != nil {
		fmt.Printf("%v", err)
	}
}

func parseRequest(s string) (req *Request, ok bool) {
	method, after1, ok1 := strings.Cut(s, " ")
	path, after2, ok2 := strings.Cut(after1, " ")
	httpversion, after3, ok3 := strings.Cut(after2, "\r\n")
	fmt.Println("after3:", after3)

	if !ok1 || !ok2 || !ok3 {
		return nil, false
	}

	rawheaders, rawbody := separateHeadersAndBody(after3)

	req = &Request{}
	req.httpversion = httpversion
	req.method = method
	req.path = path
	req.headers = mapheaders(strings.Split(rawheaders, "\r\n"))
	req.body = rawbody

	fmt.Printf("httpversion: %s\n", req.httpversion)
	fmt.Printf("method: %s\n", req.method)
	fmt.Printf("path: %s\n", req.path)
	fmt.Printf("headers: %v\n", req.headers)
	fmt.Printf("body: %v\n", req.body)
	return req, true
}

func separateHeadersAndBody(s string) (rawheaders, rawbody string) {
	split := strings.Split(s, "\r\n\r\n")

	return split[0], split[1]
}

func httpstatus(code int) string {
	switch code {
	case 200:
		return "OK"
	case 404:
		return "Not Found"
	default:
		return "OK"
	}
}

func mapheaders(ss []string) map[string]string {
	headers := make(map[string]string)
	fmt.Println(ss)
	for _, s := range ss {
		split := strings.Split(s, ": ")
		if 2 != len(split) {
			panic("ERROR: Malformed header")
		}

		headers[split[0]] = split[1]
	}

	return headers
}
