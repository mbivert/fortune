package main

import (
	"bufio"
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Page struct {
	Id      string
	Content string
}

var (
	port       = flag.String("port", "8081", "Listening HTTP port")
	fortuneDir = flag.String("dir", "./fortunes", "Fortune directory")
	fortunes   [][]byte
)

func loadFortunes(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	var fortune []byte

	line, err := r.ReadBytes('\n')
	for err == nil {
		/* next fortune */
		if len(line) == 2 && line[0] == '%' {
			fortunes = append(fortunes, fortune)
			fortune = nil
		} else {
			fortune = append(fortune, line...)
		}
		line, err = r.ReadBytes('\n')
	}

	/* add a random fortune if none */
	if len(fortunes) == 0 {
		fortunes = append(fortunes, []byte("No fortune!"))
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.URL.Path[1:])
	if err != nil {
		n = rand.Int()
	}

	if n >= len(fortunes) {
		n = n % len(fortunes)
	}
	skel, _ := template.ParseFiles("fortune.html")

	page := &Page{Id: strconv.Itoa(n), Content: string(fortunes[n])}
	skel.Execute(w, page)
}

func main() {
	rand.Seed(time.Now().Unix())

	flag.Parse()

	if files, err := filepath.Glob(*fortuneDir + "/*"); err != nil {
		log.Fatal(err)
	} else {
		for _, f := range files {
			if err := loadFortunes(f); err != nil {
				log.Println(err)
			}
		}
	}

	log.Println(len(fortunes), "fortunes loaded")

	http.HandleFunc("/", handler)

	log.Println("Launching on http://localhost:" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
