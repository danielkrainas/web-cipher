package decode

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/danielkrainas/weph/cipher"
	"github.com/danielkrainas/weph/cmd"
	"github.com/danielkrainas/weph/context"
)

func init() {
	cmd.Register("decode", Info)
}

func readUrlList(filename string) ([]string, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer fp.Close()
	s := bufio.NewScanner(fp)
	urls := make([]string, 0)
	for s.Scan() {
		urls = append(urls, s.Text())
	}

	return urls, nil
}

func readMessageFromStdin() string {
	bio := bufio.NewReader(os.Stdin)
	result := ""
	emptyCount := 0
	for {
		line, err := bio.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			emptyCount++
			if emptyCount > 1 {
				break
			}
		} else if emptyCount > 0 {
			result += "\n"
			emptyCount = 0
		}

		result += line
	}

	return result
}

func run(ctx context.Context, args []string) error {
	encodedMsg := context.GetStringValue(ctx, "flags.message")
	if readFromIn, ok := ctx.Value("flags.in").(bool); ok {
		if encodedMsg != "" && readFromIn {
			return errors.New("cannot specify a message flag and reading from input")
		} else if readFromIn {
			encodedMsg = readMessageFromStdin()
		}
	}

	if encodedMsg == "" {
		return fmt.Errorf("you must specify a message to decode")
	}

	var urls []string
	var err error
	urlListFile := context.GetStringValue(ctx, "flags.urls")
	if urlListFile != "" {
		urls, err = readUrlList(urlListFile)
		if err != nil {
			return err
		}
	}

	var references []*cipher.PageReference
	i := uint16(0)
	for _, url := range urls {
		refs, err := cipher.GetReferences(url, i)
		if err != nil {
			return fmt.Errorf("error getting references: %v\n", err)
		}

		references = append(references, refs...)
		i++
	}

	for _, glyph := range strings.Split(encodedMsg, "/") {
		e := cipher.FromBase77(glyph)
		r := cipher.Lookup(e.Reference.Index, e.Reference.Level, e.Reference.Url, references)
		if r == nil || uint16(len(r.Text)) < e.CharIndex {
			fmt.Print("#")
		} else {
			fmt.Print(r.Text[e.CharIndex])
		}
	}

	fmt.Print("\n\n")
	return nil
}

var (
	Info = &cmd.Info{
		Use:   "decode",
		Short: "`decode`",
		Long:  "`decode`",
		Run:   cmd.ExecutorFunc(run),
		Flags: []*cmd.Flag{
			{
				Short:       "u",
				Long:        "urls",
				Type:        cmd.FlagString,
				Description: "",
			},
			{
				Short:       "i",
				Long:        "in",
				Type:        cmd.FlagBool,
				Description: "",
				Default:     false,
			},
			{
				Short:       "m",
				Long:        "message",
				Type:        cmd.FlagString,
				Description: "",
			},
		},
	}
)
