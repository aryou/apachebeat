package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/singlehopllc/apachebeat/beater"
)

func main() {
	err := beat.Run("apachebeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
