package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type (
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

	_, path, proto, _ := parseRequest(string(buf))
	_, body := separateHeadersAndBody(proto)
	res := &Response{
		httpversion: "1.1",
		body:        body,
	}

	param, found := strings.CutPrefix(path, "/echo/")
	if path == "/" {
		res.code = 200
		write(conn, res)
	} else if found {
		res.code = 200
		res.body = param
		res.headers = []string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %d", len(param))}
		write(conn, res)
	} else {
		res.code = 404
		write(conn, res)
	}

}

func parseRequest(s string) (method, path, proto string, ok bool) {
	method, aftercut1, ok1 := strings.Cut(s, " ")
	path, proto, ok2 := strings.Cut(aftercut1, " ")

	if !ok1 || !ok2 {
		return "", "", "", false
	}

	return method, path, proto, true
}

func separateHeadersAndBody(s string) (rawheaders, rawbody string) {
	split := strings.Split(s, "\r\n\r\n")

	return split[0], split[1]
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
