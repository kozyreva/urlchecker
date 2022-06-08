package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type checkup struct {
	line  []byte
	check bool
}

func main() {
	in := make(chan []byte)
	out := make(chan checkup)

	go reader(in)
	go processor(in, out)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go writer(&wg, out)
	wg.Wait()
}

func reader(in chan<- []byte) {
	defer close(in)
	file, err := os.OpenFile("url.txt", os.O_RDONLY, 0755)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	r := bufio.NewReader(file)
	for line, _, err := r.ReadLine(); err == nil; line, _, err = r.ReadLine() {
		in <- line
		//log.Println("reader", string(line))
	}
}

func processor(in <-chan []byte, out chan<- checkup) {
	defer close(out)
	num := runtime.NumCPU()
	pg := sync.WaitGroup{}
	for i := 0; i < num; i++ {
		pg.Add(1)
		go checker(&pg, in, out)
	}
	pg.Wait()
}

func checker(pg *sync.WaitGroup, in <-chan []byte, out chan<- checkup) {
	defer pg.Done()
	{
		for line := range in {
			resp, err := http.Get(string(line))
			res := checkup {
				line:  line,
				check: err == nil,
			}
			if err == nil {
				resp.Body.Close()
			}
			out <- res
			//log.Println("worker", string(res.line), res.check)
		}
	}
}

func writer(wg *sync.WaitGroup, out <-chan checkup) {
	defer wg.Done()
	file, err := os.OpenFile("urlcheck.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for res := range out {
		var sb strings.Builder
		sb.Write(res.line)
		sb.WriteRune(',')
		sb.WriteString(strconv.FormatBool(res.check))
		sb.WriteRune('\n')
		_, err := w.WriteString(sb.String())
		if err != nil {
			log.Println(err)
			return
		}
		w.Flush()
	}
}
