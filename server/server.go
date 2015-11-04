// Stat server
package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/mediocregopher/radix.v2/redis"
)

type Page struct {
	Title string
	Body  []byte
}

func loadTopPage(body []byte) *Page {
	return &Page{Title: "Top IPs", Body: body}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}

func main() {

	http.HandleFunc("/", handler)
	http.HandleFunc("/top", topDestinations)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Try /top, I do not know how to handle = %q\n", req.URL.Path)
}

func topDestinations(rw http.ResponseWriter, req *http.Request) {

	client, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Problem communicating to Redis...")
		log.Fatal(err)
	} else {
		defer client.Close()
	}

	response := client.Cmd("ZREVRANGE", "popularity", "0", "-1", "WITHSCORES")

	//ZREVRANGE traffic 0 -1 WITHSCORES
	buffer := bytes.NewBufferString("<TABLE>")

	l, _ := response.List()
	for _, elemStr := range l {
		buffer.WriteString("<TR><TD>")
		buffer.WriteString(elemStr)
		buffer.WriteString("<TR><TD>")
	}
	buffer.WriteString("</TABLE>")

	p := loadTopPage(buffer.Bytes())
	//fmt.Println("PAGE body: " + string(p.Body[:]))
	renderTemplate(rw, "top", p)

}