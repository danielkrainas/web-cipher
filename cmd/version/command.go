package version

import (
	"fmt"

	"github.com/danielkrainas/wiph/cmd"
	"github.com/danielkrainas/wiph/context"
)

func init() {
	cmd.Register("version", Info)
}

func run(ctx context.Context, args []string) error {
	fmt.Println("Wiph v", context.GetVersion(ctx))
	return nil
}

var (
	Info = &cmd.Info{
		Use:   "version",
		Short: "`version`",
		Long:  "`version`",
		Run:   cmd.ExecutorFunc(run),
	}
)
