package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var httpClient = http.Client{
	Timeout: 10 * time.Second,
}

func check(url string) (*Site, error) {
	req, err := http.NewRequest("GET", "http://"+url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, errors.New("non-OK response code")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	s := Site{
		Url:   url,
		Compr: resp.Header.Get("Content-Encoding"),
		Size:  len(data),
	}

	// see how much we can save applying gzip on the response body
	buf := bytes.Buffer{}
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	s.SizeCompressed = len(buf.Bytes())
	return &s, nil
}

func checkAll() {
	sites, err := getSites()
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan Site)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for s := range ch {
				log.Printf("Checking %s\n", s.Url)
				site, err := check(s.Url)
				if err != nil {
					log.Printf("%s error: %s\n", s.Url, err)

					// mark site as checked without updating any of its properties
					_ = updateSite(&s)

					continue
				}

				if err := updateSite(site); err != nil {
					log.Printf("%s database error: %s\n", s.Url, err)
					continue
				}
			}
		}()
	}

	for _, s := range sites {
		ch <- s
	}

	close(ch)
	wg.Wait()
}

func writeResults(name string) {

	tmpl := template.Must(template.ParseFiles("template.gohtml"))
	sites, err := getSitesOrderedBySavings()
	if err != nil {
		log.Printf("error writing template: %s\n", err)
	}

	viewCtx := struct {
		Sites []Site
	}{
		Sites: sites,
	}

	fh, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	if err := tmpl.Execute(fh, viewCtx); err != nil {
		log.Printf("error writing template: %s\n", err)
	}
}

func main() {
	writeResults("sites-without-compression.html")
	checkAll()
}
