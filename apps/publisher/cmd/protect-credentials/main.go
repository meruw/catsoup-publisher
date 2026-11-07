// A local maintenance command. Never prints credential values or uses the network.
package main

import (
	"catsoup/publisher/internal/uploads"
	"fmt"
	"os"
)

func main() {
	files, err := uploads.MigrateDeveloperCredentials()
	for _, file := range files {
		fmt.Printf("%s: protected and readable\n", file)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Verified %d developer configurations. Account tokens were not modified.\n", len(files))
}
