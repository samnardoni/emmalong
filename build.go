package main

import (
	"encoding/json"
	"fmt"
	"image/jpeg"
	"net/http"
	"os"
)

const (
	maxLength = 10000
)

var (
	photosets = map[string]string{
		"dance":    "72157625050708435",
		"portrait": "72157625175466660",
		"nature":   "72157625175473042",
	}
)

type FlickrMessage struct {
	Photoset FlickrPhotoset
	Stat     string
}

type FlickrPhotoset struct {
	Id    string
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

func NewPhoto(URL string) *Photo {
	photo := &Photo{URL: URL}

	return photo
}

// This isn't called in NewPhoto because parallelizing NewPhoto
// means that a photoset is in a different order. This allows us
// to add a photo to the list in order, then download photos when we want
// to.
func (photo *Photo) SetDimensions() {

	client := &http.Client{}
	response, _ := client.Get(photo.URL)

	config, _ := jpeg.DecodeConfig(response.Body)

	photo.Width = config.Width
	photo.Height = config.Height
}

type Output map[string][]*Photo

func main() {

	client := &http.Client{}
	buffer := make([]byte, maxLength)
	output := make(Output)

	for name, id := range photosets {

		fmt.Printf("[ %s ]\n", name)

		response, _ := client.Get(FlickrPhotosetURL(id))
		n, _ := response.Body.Read(buffer)

		var m FlickrMessage

		err := json.Unmarshal(buffer[:n], &m)
		if err != nil {
			fmt.Println("Error with JSON: ", err)
			return
		}

		c := make(chan bool)

		for _, fPhoto := range m.Photoset.Photo {
			photo := NewPhoto(fPhoto.URL("b"))
			output[name] = append(output[name], photo)
		}

		// Set dimensions of photo in parallel
		for _, photo := range output[name] {
			go func(p *Photo) {
				p.SetDimensions()
				fmt.Println(p.URL)
				c <- true
			}(photo)
		}

		for _ = range output[name] {
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
	file.WriteString("var photos = ")
	file.Write(marshalled)
	file.WriteString(";\n")
}
