package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func root(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.ParseFiles("main.html"))
	tmpl.Execute(w, nil)
}

func renderManifest(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	text := req.Form.Get("manifest")
	if text != "" {
		manifestByte, err := sortManifest([]byte(text))
		if err != nil {
			fmt.Fprintf(w, "Cannot sort manifest, due to err: %v", err)
		}
		fmt.Fprintf(w, "%s", manifestByte)
	}
}

func startWebserver() {
	http.HandleFunc("/", root)
	http.HandleFunc("/render", renderManifest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}