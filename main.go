package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/common/transform/typeconv"
)

type regOp struct {
	Op string `json:"op"`
	ID int    `json:"id"`
}

type checkpointEntry struct {
	Cursor  *Cursor       `json:"cursor"`
	Key     string        `json:"_key"`
	Meta    *Meta         `json:"meta"`
	TTL     time.Duration `json:"ttl"`
	Updated time.Time     `json:"updated"`
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

type entry struct {
	K     string                 `json:"k"`
	Value map[string]interface{} `json:"v"`
}

func (e entry) Decode(to interface{}) error {
	return typeconv.Convert(to, e.Value)
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("cannot read file %q, error: %s", os.Args[1], err)
	}
	defer file.Close()

	if err := processData(file, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func processLine(line string) (string, error) {
	switch {
	case strings.Contains(line, "\"_key\""):
		l, err := decodeCheckpointEntry(line)
		if err != nil {
			return "", err
		}
		return l, nil
	case strings.Contains(line, "\"k\""):
		l, err := decodeLogEntry(line)
		if err != nil {
			return "", err
		}
		return l, nil
	case strings.Contains(line, "\"op\""):
		l, err := decodeRegOp(line)
		if err != nil {
			return "", err
		}
		return l, nil
		// the checkpoint file is a JSON array, skip the square brackets
	case line == "[" || line == "]":
		return "", nil
	default:
		return "", fmt.Errorf("unknown format: %q", line)
	}
}

func decodeCheckpointEntry(line string) (string, error) {
	if strings.HasSuffix(line, ",") {
		line = line[:len(line)-1]
	}
	m := map[string]any{}
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return "", fmt.Errorf("cannot decode line into map[string]any: %w", err)
	}

	e := checkpointEntry{}
	if err := typeconv.Convert(&e, m); err != nil {
		return "", fmt.Errorf("cannot use typeconv to decode line '%q', err: %w", line, err)
	}

	key, ok := m["_key"].(string)
	if !ok {
		return "", fmt.Errorf("'_key' is %T instead of string", m["_key"])
	}
	e.Key = key

	finalBytes, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("could not encode checkpoint entry as JSON, err: %w", err)
	}
	return string(finalBytes), nil
}

// decodeRegOp ensures line is in the correct format
func decodeRegOp(line string) (string, error) {
	op := regOp{}
	if err := json.Unmarshal([]byte(line), &op); err != nil {
		return "", fmt.Errorf("cannot decode line: %w", err)
	}

	if op.ID == 0 || op.Op == "" {
		return "", fmt.Errorf("%q is not a valid registry operation", line)
	}

	return line, nil
}

func processData(in io.Reader, out io.Writer) error {
	s := bufio.NewScanner(in)
	var lineCount uint64
	for s.Scan() {
		opLine := s.Text()
		line, err := processLine(opLine)
		if err != nil {
			return fmt.Errorf("could not process line %d, err: %w", lineCount, err)
		}

		// no need to write a blank line
		if line == "" {
			continue
		}

		if _, err := out.Write([]byte(line)); err != nil {
			return fmt.Errorf("could not write line %d, err: %w", lineCount, err)
		}
		if _, err := out.Write([]byte("\n")); err != nil {
			return fmt.Errorf("could not write line break after line %d, error: %s", lineCount, err)
		}
		lineCount++
	}

	return nil
}

func decodeLogEntry(line string) (string, error) {
	var dec entry
	var st Value

	if err := json.Unmarshal([]byte(line), &dec); err != nil {
		panic(err)
	}

	if err := dec.Decode(&st); err != nil {
		log.Fatalf("Failed to read regisry state for '%q', error: %#v", dec.K, err)
		return "", nil
	}

	ret := regEntry{
		K: dec.K,
		V: st,
	}

	humanRadable, err := json.Marshal(ret)
	if err != nil {
		return "", fmt.Errorf("could not encode registry entry as JSON: %w", err)
	}

	return string(humanRadable), nil
}
