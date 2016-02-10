package main

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
)

var (
	encodingChars = "abcdefghi=jklmnopqrst!uvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321@#$%^&*~.+_-,"
)

type pageReference struct {
	text  string
	level uint16
	index uint16
	url   uint16
}

type encodedReference struct {
	charIndex uint16
	ch        byte
	reference *pageReference
}

func main() {
	var url = os.Args[1]
	var url2 = os.Args[2]
	msg := "the quick fox jumps over the gate for whatever reason and danny dances the jig"
	references, err := getReferences(url, 0)
	if err != nil {
		fmt.Errorf("error getting references: %v\n", err)
		return
	}

	otherRefs, err := getReferences(url2, 1)
	if err != nil {
		fmt.Errorf("error getting references: %v\n", err)
		return
	}

	references = append(references, otherRefs...)

	//fmt.Printf("references: %d\n", len(references))
	var used []*encodedReference
	buf := []byte(msg)
	for _, b := range buf {
		encoded := nextReference(b, used, references)
		if encoded == nil {
			fmt.Printf("Couldn't find anything for %x\n", b)
		} else {
			//fmt.Printf("%s %s \n", string([]byte{b}), toBase77(encoded))
			fmt.Printf("%s", toBase77(encoded))
			used = append(used, encoded)
		}
	}

	//fmt.Println("")
	//fmt.Println("done")
}

func (encoded *encodedReference) toUint64() uint64 {
	if encoded == nil {
		return 0
	}

	buf := make([]byte, 8)
	encoder := binary.BigEndian
	encoder.PutUint16(buf, encoded.reference.url)
	encoder.PutUint16(buf[2:], encoded.charIndex)
	encoder.PutUint16(buf[4:], encoded.reference.level)
	encoder.PutUint16(buf[6:], encoded.reference.index)
	return encoder.Uint64(buf)
}

func toBase77(encoded *encodedReference) string {
	encodingCharsLength := len(encodingChars)
	i := encoded.toUint64()
	var s []string
	for m := 8; m > 0; m-- {
		c := uint64(math.Pow(float64(encodingCharsLength), float64(m)))
		if c < 0 {
			// special case, handle later
			continue
		}

		if i >= c {
			d := uint64(i / c)
			i -= d * c
			//fmt.Printf("%d ", d)
			s = append([]string{string(encodingChars[int(d)])}, s...)
		} else if i <= 0 {
			break
		}
	}

	if i > 0 {
		s = append([]string{string(encodingChars[int(i)])}, s...)
	} else if len(s) > 0 {
		s = append([]string{string(encodingChars[0])}, s...)
	}

	return strings.Join(s, "")
}

func downloadContents(url string) (io.ReadCloser, error) {
	//fmt.Printf("fetching url: %s\n", url)
	res, err := http.Get(url)
	return res.Body, err
}

func findMatches(c byte, references []*pageReference) []*pageReference {
	var matches []*pageReference
	for _, r := range references {
		if strings.IndexByte(r.text, c) > -1 {
			matches = append(matches, r)
		}
	}

	return matches
}

func getLastUsedFor(c byte, ref *pageReference, used []*encodedReference) *encodedReference {
	for i := len(used) - 1; i >= 0; i-- {
		u := used[i]
		if u.reference.index == ref.index && u.reference.url == ref.url && u.reference.level == ref.level && u.ch == c {
			return u
		}
	}

	return nil
}

func nextReference(c byte, used []*encodedReference, references []*pageReference) *encodedReference {
	matches := findMatches(c, references)
	for _, r := range matches {
		start := 0
		if previous := getLastUsedFor(c, r, used); previous != nil {
			start = int(previous.charIndex + 1)
		}

		if start >= len(r.text) {
			continue
		}

		index := strings.IndexByte(r.text[start:], c)
		if index >= 0 {
			encoded := &encodedReference{
				reference: r,
				ch:        c,
				charIndex: uint16(index),
			}

			return encoded
		}
	}

	return nil
}

func getReferences(url string, index uint16) ([]*pageReference, error) {
	res, err := downloadContents(url)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	var references []*pageReference
	var depth uint16 = 0
	var counter uint16 = 0
	tokenizer := html.NewTokenizer(res)

	done := false
	ignoreContents := false
	for !done {
		tokenType := tokenizer.Next()

		switch {
		case tokenType == html.ErrorToken:
			done = true
			break

		case tokenType == html.TextToken:
			if ignoreContents {
				continue
			}

			ref := &pageReference{
				url:   index,
				index: counter,
				level: depth,
			}

			ref.text = string(tokenizer.Text()[:])
			references = append(references, ref)
			counter++
			continue

		case tokenType == html.StartTagToken:
			depth++
			token := tokenizer.Token()
			if token.Data == "script" || token.Data == "style" {
				ignoreContents = true
			}

			continue

		case tokenType == html.EndTagToken:
			ignoreContents = false
			depth--
			continue
		}
	}

	return references, nil
}
