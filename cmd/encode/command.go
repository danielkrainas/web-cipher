package encode

import (
	"fmt"

	"github.com/danielkrainas/wiph/cipher"
	"github.com/danielkrainas/wiph/cmd"
	"github.com/danielkrainas/wiph/context"
)

func init() {
	cmd.Register("encode", Info)
}

func run(ctx context.Context, args []string) error {
	urs := context.GetStringValue(ctx, "flags.urls")
	var url = args[0]
	var url2 = args[1]
	msg := context.GetStringValue(ctx, "flags.message")
	if msg == "" {
		msg = "the quick fox jumps over the gate for whatever reason and danny dances the jig"
	}

	references, err := cipher.GetReferences(url, 0)
	if err != nil {
		return fmt.Errorf("error getting references: %v\n", err)
	}

	otherRefs, err := cipher.GetReferences(url2, 1)
	if err != nil {
		return fmt.Errorf("error getting references: %v\n", err)
	}

	references = append(references, otherRefs...)

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
		},
	}
)
