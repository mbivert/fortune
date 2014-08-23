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
	skel, _    = template.ParseFiles("fortune.html")
	fortunes   []string
	addc       = make(chan string)
	idc        = make(chan int)
)

func loadFortunes(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	var fortune []byte

	line, err := r.ReadString('\n')
	for err == nil {
		/* next fortune */
		if len(line) == 2 && line[0] == '%' {
			fortunes = append(fortunes, string(fortune))
			fortune = nil
		} else {
			fortune = append(fortune, line...)
		}
		line, err = r.ReadString('\n')
	}

	/* add a random fortune if none */
	if len(fortunes) == 0 {
		fortunes = append(fortunes, "No fortune!")
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

	if r.FormValue("raw") == "" {
		page := &Page{Id: strconv.Itoa(n), Content: fortunes[n]}
		skel.Execute(w, page)
	} else {
		w.Write([]byte(fortunes[n]))
	}
}

func addUser(fortune string) {
	f, err := os.OpenFile(*fortuneDir+"/ufortunes",
		os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
	} else {
		n, err := f.WriteString(fortune)
		if err != nil { log.Println(err, n) }
		f.WriteString("\n%\n")
		f.Close()
	}
}

func userFortunes() {
	for {
		select {
		case fortune := <- addc:
			fortunes = append(fortunes, fortune)
			addUser(fortune)
			idc <- len(fortunes)-1
		}
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	raw := ""
	// raw anti-spam
	if r.FormValue("content") != "" { return }

	if fortune := r.FormValue("fortune"); fortune != "" {
		addc <- fortune
	}

	if r.FormValue("raw") != "" { raw = "?raw=1" }

	http.Redirect(w, r, "/"+strconv.Itoa(<- idc)+raw, http.StatusFound)
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

	go userFortunes()

	log.Println(len(fortunes), "fortunes loaded")

	http.HandleFunc("/", handler)
	http.HandleFunc("/add", add)

	log.Println("Launching on http://localhost:" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
