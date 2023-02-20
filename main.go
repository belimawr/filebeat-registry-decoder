package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("cannot read file %q, error: %s", os.Args[1], err)
	}
	defer file.Close()

	// out, err := os.Create(os.Args[2])
	// if err != nil {
	// 	log.Fatalf("cannot create output file %q, error: %s", os.Args[2], err)
	// }
	// defer out.Close()

	processData(file, os.Stdout)
}

func processData(in io.Reader, out io.Writer) error {
	s := bufio.NewScanner(in)
	var line uint64
	for s.Scan() {
		// First line is always an "operation"
		// so just print it again
		line++
		opLine := s.Bytes()
		out.Write(opLine)

		out.Write([]byte("\n"))
		// Second line is always an entry
		// decode and "pretty print" it

		if !s.Scan() {
			return nil
		}

		line++
		entryLine := s.Bytes()
		v, err := f(entryLine)
		if err != nil {
			panic(err)
		}

		d, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		out.Write(d)

		if _, err := out.Write([]byte("\n")); err != nil {
			return fmt.Errorf("could not write line break after line %d, error: %s", line, err)
		}
	}

	return nil
}

type regOp struct {
	Op string `json:"op"`
	ID int    `json:"id"`
}

type regEntry struct {
	K string `json:"k"`
	V Value  `json:"v"`
}

type Cursor struct {
	Offset int `json:"offset"`
}

type Meta struct {
	Source string `json:"source"`
	// It needs the `_` because of the decoding library.
	Identifier_Name string `json:"identifier_name"`
}

type Value struct {
	Cursor  *Cursor   `json:"cursor"`
	Meta    *Meta     `json:"meta"`
	TTL     int64     `json:"ttl"`
	Updated time.Time `json:"updated"`
}
