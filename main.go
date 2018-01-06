package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8007", "http service address")
var serveFiles = flag.Bool("serveFiles", true, "use this app to serve static files (js, css, images)")

func serveIndexPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "html/index.html")
}

func main() {
	flag.Parse()
	lobby := newLobby()
	go lobby.run()
	http.HandleFunc("/", serveIndexPage)
	if *serveFiles {
		http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./html/js"))))
		http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./html/css"))))
		http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./html/img"))))
	}
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(lobby, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
