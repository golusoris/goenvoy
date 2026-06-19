package kavita_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/mediaserver/kavita/v2"
)

func Example() {
	c, err := kavita.New("http://localhost:5000", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}
	libs, err := c.GetLibraries(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(libs))
}
