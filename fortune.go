package main
import (
	"bufio"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Page struct {
	Id      string
	Content string
}

var fortunes [][]byte

func loadfortunes(fn string) error {
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

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	if len(os.Args) <= 1 {
		err := loadfortunes("fortunes/9fortunes")
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	for i := 1; i < len(os.Args); i++ {
		err := loadfortunes(os.Args[i])
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	fmt.Printf("%d fortunes loaded\n", len(fortunes))

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8081", nil)
}
