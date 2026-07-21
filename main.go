package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/li-qs/mallard/collector"
)

func main() {
	addr := ":8090"

	args := os.Args
	if len(args) > 1 {
		addr = ":" + args[1]
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheck)

	fs := http.FileServer(http.Dir("./html"))
	mux.Handle("/ui/", http.StripPrefix("/ui/", fs))

	c := collector.NewCollector()
	mux.HandleFunc("/collect", c.HandleCollect)
	mux.HandleFunc("/trace/", c.HandleQuery)

	fmt.Printf("Collector running on %s\n", addr)
	http.ListenAndServe(addr, mux)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
