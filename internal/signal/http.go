// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package signal

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// HTTPSDPServer starts a HTTP Server that consumes SDPs
func HTTPSDPServer() chan string {
	port := flag.Int("port", 8080, "http server port")
	flag.Parse()

	sdpChan := make(chan string)
	http.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "done")
		sdpChan <- string(body)
	})

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(*port), nil) // nolint:gosec
		if err != nil {
			panic(err) //nolint
		}
	}()

	return sdpChan
}
