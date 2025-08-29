package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/devasherr/tcp-http/internal/headers"
	"github.com/devasherr/tcp-http/internal/request"
	"github.com/devasherr/tcp-http/internal/response"
	"github.com/devasherr/tcp-http/internal/server"
)

const port = 42069

func response400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func response500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func response200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func toStr(data [32]byte) string {
	res := ""
	for _, d := range data {
		res += fmt.Sprintf("%02x", d)
	}
	return res
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		status := response.StatusOk
		body := response200()

		if req.RequestLine.RequestTarget == "/yourproblem" {
			status = response.StatusBadRequest
			body = response400()
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			status = response.StatusInternalServerError
			body = response500()
		} else if req.RequestLine.RequestTarget == "/video" {
			f, _ := os.ReadFile("assets/vim.mp4")
			h.Replace("Content-type", "video/mp4")
			h.Replace("Content-length", fmt.Sprintf("%d", len(f)))

			w.WriteStatusLine(response.StatusOk)
			w.WriteHeaders(h)
			w.WriteBody(f)

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			target := req.RequestLine.RequestTarget[len("/httpbin/"):]
			resp, err := http.Get("https://httpbin.org/" + target)
			if err != nil {
				status = response.StatusInternalServerError
				body = response500()
			} else {
				w.WriteStatusLine(response.StatusOk)
				h.Delete("Content-length")
				h.Set("transfer-encoding", "chuncked")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				h.Replace("Content-type", "text/plain")
				w.WriteHeaders(h)

				fullBody := make([]byte, 0)
				for {
					data := make([]byte, 32)
					n, err := resp.Body.Read(data)
					if err != nil {
						break
					}

					fullBody = append(fullBody, data[:n]...)

					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", len(data))))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))
				// bruh, trailers are headers !!
				trailers := headers.NewHeaders()
				hash := sha256.Sum256(fullBody)
				trailers.Set("X-Content-SHA256", toStr(hash))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

				w.WriteHeaders(trailers)
				w.WriteBody([]byte("\r\n"))
				return
			}

		}

		h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-type", "text/html")

		w.WriteStatusLine(status)
		w.WriteHeaders(h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
