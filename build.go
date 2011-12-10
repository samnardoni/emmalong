package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"strconv"
	"os"
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

type Message struct {
	Photoset Photoset
	Stat     string
}

type Photoset struct {
	Id     string
	Photo []Photo
}

type Photo struct {
	Id     string
	Farm   float64
	Server string
	Secret string
	Title  string
}

type URLs []string
type Output map[string]URLs

func (p *Photo) URL(size string) string {
	//http://farm{farm-id}.staticflickr.com/{server-id}/{id}_{secret}_[mstzb].jpg

	return "http://farm" + strconv.Ftoa64(p.Farm, 'f', 0) + 
		".staticflickr.com/" + p.Server + "/" + p.Id + "_" + p.Secret + "_" + 
		size + ".jpg"
}

func main() {

	client := &http.Client {}
	buffer := make([]byte, maxLength)
	output := make(Output)

	for photosetName, photosetId := range(photosets) {

		response, _ := client.Get(FlickrPhotosetURL(photosetId))
		n, _ := response.Body.Read(buffer)

		var m Message

		err := json.Unmarshal(buffer[:n], &m)
		if err != nil {
			fmt.Println("Error with JSON: ", err)
			return
		}

		for _, photo := range(m.Photoset.Photo) {
			output[photosetName] = append(output[photosetName], photo.URL("b"))
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