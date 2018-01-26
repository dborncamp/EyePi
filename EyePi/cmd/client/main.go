package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Image struct {
	ImageName string `string`
	Image     []byte `json:"image"`
}

func main() {
	maindir := "/home/pi/timelapseImages"
	for i := 0; i < 30; i++ {

		filename := fmt.Sprintf("%s%d_image.jpg", maindir, i)
		req, err := transferImage(filename)
		if err != nil {
			log.Fatal(err)
		}
		//os.Remove(filename)
		fmt.Println(req)
	}
}

func transferImage(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
	}

	i := Image{
		ImageName: filename,
		Image:     b,
	}

	obj, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	log.Println(obj)
	//req, err := http.NewRequest("POST", "http://10.1.10.116:8080/image", bytes.NewBuffer(obj))
	if err != nil {
		return "", err
	}

	//res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	//log.Println(res.StatusCode)
	return "", nil
}
