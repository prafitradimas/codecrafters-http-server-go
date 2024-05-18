package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

type (
	Request struct {
		httpversion string
		host        string
		method      string
		path        string
		headers     map[string]string
		body        []byte
	}

	Response struct {
		httpversion string
		code        int
		status      string
		headers     []string
		body        []byte
	}
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buf := make([]byte, 4096)
	_, err := conn.Read(buf)
	if err != nil {
		return
	}

	req, _ := parseRequest(string(buf))
	res := &Response{
		httpversion: "1.1",
		body:        req.body,
	}

	if req.path == "/" {
		res.code = 200
		write(conn, res)
	} else if req.path == "/user-agent" {
		res.code = 200
		res.body = []byte(req.headers["User-Agent"])
		res.headers = []string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %d", len(res.body))}
		write(conn, res)
	} else if strings.HasPrefix(req.path, "/echo/") {
		param, _ := strings.CutPrefix(req.path, "/echo/")

		res.code = 200
		res.body = []byte(param)
		res.headers = []string{"Content-Type: text/plain", fmt.Sprintf("Content-Length: %d", len(res.body))}
		write(conn, res)
	} else if strings.HasPrefix(req.path, "/files/") && req.method == "GET" {
		dir := os.Args[2]
		filename := strings.TrimPrefix(req.path, "/files/")
		buf, err := os.ReadFile(path.Join(dir, filename))
		if err != nil {
			res.code = 404
			write(conn, res)
		} else {
			res.code = 200
			res.body = buf
			res.headers = []string{"Content-Type: application/octet-stream", fmt.Sprintf("Content-Length: %d", len(res.body))}
			write(conn, res)
		}
	} else if strings.HasPrefix(req.path, "/files/") && req.method == "POST" {
		dir := os.Args[2]
		filename := strings.TrimPrefix(req.path, "/files/")
		err := os.WriteFile(path.Join(dir, filename), bytes.Trim(req.body, "\x00"), 0644)
		if err != nil {
			res.code = 404
			write(conn, res)
		} else {
			res.code = 201
			res.body = []byte("")
			res.headers = []string{}
			write(conn, res)
		}
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

	if !ok1 || !ok2 || !ok3 {
		return nil, false
	}

	rawheaders, rawbody := separateHeadersAndBody(after3)

	req = &Request{}
	req.httpversion = httpversion
	req.method = method
	req.path = path
	req.headers = mapheaders(strings.Split(rawheaders, "\r\n"))
	req.body = []byte(rawbody)

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
	case 201:
		return "Created"
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
