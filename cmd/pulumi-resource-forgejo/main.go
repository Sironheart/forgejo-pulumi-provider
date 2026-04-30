package main

import (
	"context"
	"fmt"
	"os"

	forgejoprovider "forgejo.siron.casa/sironheart/forgejo-pulumi-provider/internal/provider"
)

var version = "0.0.0-dev"

func main() {
	if err := forgejoprovider.Provider().Run(context.Background(), "forgejo", version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
