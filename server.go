package main

import (
	"log"
	"net/http"
)

func main() {

	fs := http.FileServer(http.Dir("output"))
	http.Handle("/", fs)
	log.Println("Serving Crisp Blog on Port 3000 ...")
	http.ListenAndServe(":3000", nil)

}