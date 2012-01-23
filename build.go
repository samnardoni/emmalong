package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"os"
	"image/jpeg"
)

const (
	maxLength = 10000
)

var (
	photosets = map[string]string {
		"people":    "72157625050708435",
		"urban":     "72157625050716803",
		"landscape": "72157625175466660",
		"nature":    "72157625175473042",
		"misc":      "72157625050738565",
	}
)

type FlickrMessage struct {
	Photoset FlickrPhotoset
	Stat     string
}

type FlickrPhotoset struct {
	Id     string
	Photo []FlickrPhoto
}

type FlickrPhoto struct {
	Id     string
	Farm   float64
	Server string
	Secret string
	Title  string
}

func (p *FlickrPhoto) URL(size string) string {
	return fmt.Sprintf("http://farm%d.staticflickr.com/%s/%s_%s_%s.jpg",
						uint64(p.Farm), p.Server, p.Id, p.Secret, size)
}

type Photo struct {
	URL    string
	Width  int
	Height int	
}

func (p *Photo) SetWidthAndHeight() {
	fmt.Println(" - ", p.URL, "...")

	client := &http.Client {}
	response, _ := client.Get(p.URL)

	config, _ := jpeg.DecodeConfig(response.Body)

	p.Width = config.Width
	p.Height = config.Height
}

type Output map[string][]*Photo

func main() {

	client := &http.Client {}
	buffer := make([]byte, maxLength)
	output := make(Output)

	for photosetName, photosetId := range(photosets) {

		fmt.Printf("[ %s ]\n", photosetName)

		response, _ := client.Get(FlickrPhotosetURL(photosetId))
		n, _ := response.Body.Read(buffer)

		var m FlickrMessage

		err := json.Unmarshal(buffer[:n], &m)
		if err != nil {
			fmt.Println("Error with JSON: ", err)
			return
		}

		c := make(chan bool)

		for _, flickrPhoto := range(m.Photoset.Photo) {
			photo := &Photo {URL: flickrPhoto.URL("b")}
			output[photosetName] = append(output[photosetName], photo)

			go func() {
				photo.SetWidthAndHeight()
				c <- true
			}()
		}

		for i := 0; i < len(m.Photoset.Photo); i++ {
			<-c
		}

	}

	Save("javascripts/photos.js", output)
	
}

func FlickrPhotosetURL(photoset string) string {
	return "http://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=a7a41add565d37c3ea0790ace3faf1f5&photoset_id=" + photoset + "&format=json&nojsoncallback=1"
}

func Save(filename string, output Output) {
	marshalled, _ := json.Marshal(output)
	file, _ := os.Create(filename)
	file.WriteString("var photos = ");
	file.Write(marshalled)
	file.WriteString(";\n")
}