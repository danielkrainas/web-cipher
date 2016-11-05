package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/danielkrainas/weph/cmd"
	_ "github.com/danielkrainas/weph/cmd/decode"
	_ "github.com/danielkrainas/weph/cmd/encode"
	"github.com/danielkrainas/weph/cmd/root"
	_ "github.com/danielkrainas/weph/cmd/version"
	"github.com/danielkrainas/weph/context"
)

var appVersion string

const DEFAULT_VERSION = "0.0.0-dev"

func main() {
	if appVersion == "" {
		appVersion = DEFAULT_VERSION
	}

	rand.Seed(time.Now().Unix())
	ctx := context.WithVersion(context.Background(), appVersion)

	dispatch := cmd.CreateDispatcher(ctx, root.Info)
	if err := dispatch(); err != nil {
		fmt.Printf("FATAL: %v\n", err)
	}
}
