package cipher

import (
	"encoding/binary"
	"golang.org/x/net/html"
	"io"
	"math"
	"net/http"
	"strings"
)

var (
	encodingChars = []byte("abcdefghi=jklmnopqrst!uvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321@#$%^&*~.+_-,")
)

type PageReference struct {
	Text  string
	Level uint16
	Index uint16
	Url   uint16
}

type EncodedReference struct {
	CharIndex uint16
	Ch        byte
	Reference *PageReference
}

func (encoded *EncodedReference) Uint64() uint64 {
	if encoded == nil {
		return 0
	}

	buf := make([]byte, 8)
	encoder := binary.BigEndian
	encoder.PutUint16(buf, encoded.Reference.Url)
	encoder.PutUint16(buf[2:], encoded.CharIndex)
	encoder.PutUint16(buf[4:], encoded.Reference.Level)
	encoder.PutUint16(buf[6:], encoded.Reference.Index)
	return encoder.Uint64(buf)
}

func (e *EncodedReference) Base77() string {
	encodingCharsLength := len(encodingChars)
	i := e.Uint64()
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

func indexOfEncodingChar(a byte) int {
	for i, b := range encodingChars {
		if b == a {
			return i
		}
	}

	return 0
}

func ToBase10(buf []byte) uint64 {
	j := uint64(0)
	k := []uint64{1, 77, 5929, 456533, 35153041, 2706784157}
	if len(buf) > 5 {
		j += k[5] * uint64(indexOfEncodingChar(buf[5]))
	}

	if len(buf) > 4 {
		j += k[4] * uint64(indexOfEncodingChar(buf[4]))
	}

	if len(buf) > 3 {
		j += k[3] * uint64(indexOfEncodingChar(buf[3]))
	}

	if len(buf) > 2 {
		j += k[2] * uint64(indexOfEncodingChar(buf[2]))
	}

	if len(buf) > 1 {
		j += k[1] * uint64(indexOfEncodingChar(buf[1]))
	}

	if len(buf) >= 1 {
		j += k[0] * uint64(indexOfEncodingChar(buf[0]))
	}

	return j
}

func FromBase77(s string) *EncodedReference {
	i := ToBase10([]byte(s))
	buf := make([]byte, 8)
	encoder := binary.BigEndian
	encoder.PutUint64(buf, i)
	e := &EncodedReference{
		CharIndex: encoder.Uint16(buf[2:]),
		Reference: &PageReference{
			Url:   encoder.Uint16(buf),
			Level: encoder.Uint16(buf[4:]),
			Index: encoder.Uint16(buf[6:]),
		},
	}

	return e
}

func Lookup(index uint16, level uint16, siteIndex uint16, references []*PageReference) *PageReference {
	for _, r := range references {
		if r.Index == index && r.Level == level && r.Url == siteIndex {
			return r
		}
	}

	return nil
}

func downloadContents(url string) (io.ReadCloser, error) {
	//fmt.Printf("fetching url: %s\n", url)
	res, err := http.Get(url)
	return res.Body, err
}

func findMatches(c byte, references []*PageReference) []*PageReference {
	var matches []*PageReference
	for _, r := range references {
		if strings.IndexByte(r.Text, c) > -1 {
			matches = append(matches, r)
		}
	}

	return matches
}

func getLastUsedFor(c byte, ref *PageReference, used []*EncodedReference) *EncodedReference {
	for i := len(used) - 1; i >= 0; i-- {
		u := used[i]
		if u.Reference.Index == ref.Index && u.Reference.Url == ref.Url && u.Reference.Level == ref.Level && u.Ch == c {
			return u
		}
	}

	return nil
}

func NextReference(c byte, used []*EncodedReference, references []*PageReference) *EncodedReference {
	matches := findMatches(c, references)
	for _, r := range matches {
		start := 0
		if previous := getLastUsedFor(c, r, used); previous != nil {
			start = int(previous.CharIndex + 1)
		}

		if start >= len(r.Text) {
			continue
		}

		index := strings.IndexByte(r.Text[start:], c)
		if index >= 0 {
			encoded := &EncodedReference{
				Reference: r,
				Ch:        c,
				CharIndex: uint16(index),
			}

			return encoded
		}
	}

	return nil
}

func GetReferences(url string, index uint16) ([]*PageReference, error) {
	res, err := downloadContents(url)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	var references []*PageReference
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

			ref := &PageReference{
				Url:   index,
				Index: counter,
				Level: depth,
			}

			ref.Text = string(tokenizer.Text()[:])
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
