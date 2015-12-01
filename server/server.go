// Stat server
package main

import (
	"bytes"
	"encoding/json"
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

func main() {
	// register endpoint handlers here
	http.HandleFunc("/", handler)
	http.HandleFunc("/top", topDestinations)
	http.HandleFunc("/total", jsonTotal)
	http.HandleFunc("/topip", jsonTop)
	http.HandleFunc("/botip", jsonBot)

	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

// JSON endpoint hanlders, note that we create a new REDIS connection everytime because
// the handler is ran in its own goroutine (thread), but Redis is designed to make the connection
// creation very fast, so there is not much overhead

// returns the bottom x addresses, responds to GET <statserver>/botip?limit=x
func jsonBot(w http.ResponseWriter, req *http.Request) {
	number := req.FormValue("limit") //read the limit from the request

	client, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	} else {
		defer client.Close()
	}

	response := client.Cmd("ZRANGEBYSCORE", "popularity", "-inf", "+inf", "WITHSCORES", "LIMIT", "0", number)
	list, _ := response.List()

	encoder := json.NewEncoder(w)
	encoder.Encode(list)
}

// returns the top x addresses, responds to GET <statserver>/topip?limit=x
func jsonTop(w http.ResponseWriter, req *http.Request) {

	number := req.FormValue("limit") // read the limit from the request
	client, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	} else {
		defer client.Close()
	}

	response := client.Cmd("ZREVRANGEBYSCORE", "popularity", "+inf", "-inf", "WITHSCORES", "LIMIT", "0", number)
	list, _ := response.List()

	encoder := json.NewEncoder(w)
	encoder.Encode(list)
}

// JSON end point for total number of IPs being tracked, responds to GET <statserver>/total
func jsonTotal(w http.ResponseWriter, req *http.Request) {

	client, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	} else {
		defer client.Close()
	}

	response := client.Cmd("ZCARD", "popularity")
	count, _ := response.Int() // get the count value
	// get a json encoder
	encoder := json.NewEncoder(w)
	// create a map with key and value that json encoder will respond with
	n := map[string]int{"count": count}
	encoder.Encode(n)

}

// http (html rendering) handlers
func loadTopPage(body []byte) *Page {
	return &Page{Title: "Top IPs", Body: body}
}

// renders the <name>.html template, templates should be deployed with the executables
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}

func handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Try /I do not know how to handle = %q\n", req.URL.Path)
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
	// here we create a byte buffer to construct the html response
	buffer := bytes.NewBufferString("<TABLE>")

	l, _ := response.List()
	for _, elemStr := range l {
		buffer.WriteString("<TR><TD>")
		buffer.WriteString(elemStr)
		buffer.WriteString("<TR><TD>")
	}
	buffer.WriteString("</TABLE>")

	p := loadTopPage(buffer.Bytes())
	renderTemplate(rw, "top", p)

}
