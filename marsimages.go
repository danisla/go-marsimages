package marsimages

import (
	"encoding/json"
	"fmt"
	"net/http"

	"log"

	errors "github.com/go-openapi/errors"
)

// ImageManifest from https://mars.jpl.nasa.gov/msl-raw-images/image/image_manifest.json
type ImageManifest struct {
	LatestSol int64         `json:"latest_sol"`
	Sols      []ManifestSol `json:"sols"`
	NumImages int64         `json:"num_images"`
}

// ManifestSol items in image manifest
type ManifestSol struct {
	Sol         int64  `json:"sol"`
	NumImages   int64  `json:"num_images"`
	CatalogURL  string `json:"catalog_url"`
	LastUpdated string `josn:"last_updated"`
}

// SolCatalog items from CatalogURL
type SolCatalog struct {
	Sol    int64      `json:"sol"`
	Images []SolImage `json:"images"`
}

// SolImage from catalog url
type SolImage struct {
	Sol        string `json:"sol"`
	Instrument string `json:"instrument"`
	URL        string `json:"urlList"`
	LMST       string `json:"lmst"`
	UTC        string `json:"utc"`
	SampleType string `json:"sampleType"`
	ItemName   string `json:"itemName"`
}

// ListOfImages data structure for MarsImage lists
type ListOfImages struct {
	Images []MarsImage `json:"images"`
}

// MarsImage data structure, normalized with custom meta.
type MarsImage struct {
	ItemName   string `json:"itemName"`
	URL        string `json:"url"`
	Instrument string `json:"instrument"`
	Sol        int64  `json:"sol"`
	LMST       string `json:"lmst"`
	UTC        string `json:"utc"`
}

var imageCache ListOfImages

// FetchManifest downloads the manifest at the given baseURL
func FetchManifest(baseURL string) (ImageManifest, error) {
	var manifest ImageManifest

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return manifest, errors.New(500, fmt.Sprintf("Error building http request: %s", err))
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return manifest, errors.New(500, fmt.Sprintf("Error making client request: %s", err))
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		// TODO: need to test this
		return manifest, errors.New(500, fmt.Sprintf("Error decoding JSON response for url: %s", baseURL))
	}

	return manifest, nil
}

// FetchCatalog downloads the image catalog at hte given catalogURL
func FetchCatalog(catalogURL string) (SolCatalog, error) {

	var catalog SolCatalog

	req, err := http.NewRequest("GET", catalogURL, nil)
	if err != nil {
		return catalog, errors.New(500, fmt.Sprintf("Error building http request: %s", err))
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return catalog, errors.New(500, fmt.Sprintf("Error making client request: %s", err))
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		// TODO: need to test this
		return catalog, errors.New(500, fmt.Sprintf("Error decoding JSON response for url: %s", catalogURL))
	}

	return catalog, nil
}

// CacheLatest caches the given manifest and number of sols in memory
func CacheLatest(manifest *ImageManifest, sols int) (bool, error) {
	total := len(manifest.Sols)

	for i := total - sols; i < total; i++ {
		solImages := 0
		url := manifest.Sols[i].CatalogURL
		catalog, err := FetchCatalog(url)
		if err != nil {
			log.Printf("Error fetching url: %s: %s\n", url, err)
		} else {
			for j := 0; j < len(catalog.Images); j++ {
				img := catalog.Images[j]
				if img.SampleType != "thumbnail" {
					solImages++
					marsImage := MarsImage{ItemName: img.ItemName, URL: img.URL, Instrument: img.Instrument, LMST: img.LMST, UTC: img.UTC}
					imageCache.Images = append(imageCache.Images, marsImage)
				}
			}
			log.Printf("Found %d/%d full size images for sol %d", solImages, len(catalog.Images), catalog.Sol)
		}
	}
	return true, nil
}

// GetLatest returns a the latest list of images limited by the value of count
func GetLatest(manifest *ImageManifest, count int) (ListOfImages, error) {
	var loi ListOfImages

	loi.Images = imageCache.Images[:count]

	return loi, nil
}
