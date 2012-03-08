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

	client := &http.Client{}
	response, _ := client.Get(photo.URL)

	config, _ := jpeg.DecodeConfig(response.Body)

	photo.Width = config.Width
	photo.Height = config.Height

	return photo
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
			go func(fPhoto FlickrPhoto) {
				photo := NewPhoto(fPhoto.URL("b"))
				output[name] = append(output[name], photo)
				fmt.Println(photo.URL)
				c <- true
			}(fPhoto)
		}

		for _ = range m.Photoset.Photo {
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
