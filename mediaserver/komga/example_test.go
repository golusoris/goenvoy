package komga_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/mediaserver/komga/v2"
)

func Example() {
	c, err := komga.New("http://localhost:25600", "admin@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	libs, err := c.GetLibraries(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Libraries: %d\n", len(libs))
}
