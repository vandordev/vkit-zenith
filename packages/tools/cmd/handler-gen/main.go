package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vandordev/vkit-zenith/packages/tools/internal/handlergen"
)

func main() {
	app := flag.String("app", "", "Repository-relative service directory")
	flag.Parse()
	if err := handlergen.Generate(*app); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
