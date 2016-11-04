package main

import (
	"math/rand"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/danielkrainas/wiph/cmd"
	_ "github.com/danielkrainas/wiph/cmd/encode"
	"github.com/danielkrainas/wiph/cmd/root"
	_ "github.com/danielkrainas/wiph/cmd/version"
	"github.com/danielkrainas/wiph/context"
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
		log.Fatalln(err)
	}
}
