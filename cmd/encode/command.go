package encode

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
	cmd.Register("encode", Info)
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
	msg := context.GetStringValue(ctx, "flags.message")
	if readFromIn, ok := ctx.Value("flags.in").(bool); ok {
		if msg != "" {
			return errors.New("cannot specify a message flag and reading from input")
		} else if readFromIn {
			msg = readMessageFromStdin()
		}
	}

	if msg == "" {
		return fmt.Errorf("you must specify a message to encode")
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

	var used []*cipher.EncodedReference
	buf := []byte(msg)
	for _, b := range buf {
		encoded := cipher.NextReference(b, used, references)
		if encoded == nil {
			fmt.Printf("Couldn't find anything for %x\n", b)
		} else {
			fmt.Printf("%s", encoded.Base77())
			used = append(used, encoded)
		}
	}

	fmt.Print("\n\n")
	return nil
}

var (
	Info = &cmd.Info{
		Use:   "encode",
		Short: "`encode`",
		Long:  "`encode`",
		Run:   cmd.ExecutorFunc(run),
		Flags: []*cmd.Flag{
			{
				Short:       "u",
				Long:        "urls",
				Type:        cmd.FlagString,
				Description: "",
			},
			{
				Short:       "m",
				Long:        "message",
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
		},
	}
)
