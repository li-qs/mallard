package main

import (
	"fmt"
	"net/http"

	"github.com/li-qs/mallard/collector"
)

func main() {
	addr := ":8090"
	c := collector.NewCollector()
	http.HandleFunc("/collect", c.HandleCollect)
	http.HandleFunc("/trace/", c.HandleQuery)
	fmt.Printf("Collector running on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
