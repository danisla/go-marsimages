# Mars Rover Curiosity Images Library

A GoLang library to download the latest Mars Rover Curiosity images.

## Usage

```go
package main

import (
	"log"
	"net/http"

	marsimages "github.com/danisla/go-marsimages"
)

func main() {

	const manifestURL = "https://mars.jpl.nasa.gov/msl-raw-images/image/image_manifest.json"

	client := &http.Client{}

	manifest, err := marsimages.FetchManifest(manifestURL, client)
	if err != nil {
		log.Fatal(err)
	}

	// Cache the last 10 sols
	marsimages.CacheLatest(&manifest, 10, client)

	loi, err := marsimages.GetLatest(&manifest, 20)
	if err != nil {
		log.Fatal("Error fetching latest images.")
	}

	log.Println(loi.Images[len(loi.Images)-1])
}
```