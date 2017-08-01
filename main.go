package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	t := time.Now()
	r, _ := os.Open("test.txt")
	start := []byte(`rel="bookmark">`)
	end := []byte(`</a>`)
	b := make([]byte, len(start)*2+2)
	capture := make([]byte, 1000)
	captureLen := 0
	capturing := false
	for {
		n, err := r.Read(b)
		// fmt.Printf("n = %v err = %v b = %v\n", n, err, b)
		// fmt.Printf("b[:n] = %q\n", b[:n])
		if err == io.EOF {
			break
		}
		if !capturing {
			index := bytes.Index(b, start)
			if index > -1 {
				capturing = true
				for _, by := range b[index+len(start) : n] {
					capture[captureLen] = by
					captureLen++
				}
			}
		} else {
			index := bytes.Index(b, end)
			if index > -1 {
				capturing = false
				for _, by := range b[:index] {
					capture[captureLen] = by
					captureLen++
				}
				break
			} else {
				for _, by := range b {
					capture[captureLen] = by
					captureLen++
				}
			}
		}
	}
	r.Close()
	fmt.Println(string(capture[:captureLen]))
	fmt.Println(time.Since(t))
}
