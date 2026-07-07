package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./website"))
	log.Println("go-musicfox 官网 → http://localhost:8877")
	log.Fatal(http.ListenAndServe(":8877", fs))
}
