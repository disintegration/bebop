package main

import (
	"image/png"
	"log"
	"os"
	"unicode/utf8"

	"github.com/disintegration/letteravatar"
)

var names = []string{
	"Alice",
	"Bob",
	"Carol",
	"Dave",
	"Eve",
	"Frank",
	"Gloria",
	"Henry",
	"Isabella",
	"James",
	"Жозефина",
	"Ярослав",
}

func main() {
	for _, name := range names {
		firstLetter, _ := utf8.DecodeRuneInString(name)

		img, err := letteravatar.Draw(75, firstLetter, nil)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Create(name + ".png")
		if err != nil {
			log.Fatal(err)
		}

		err = png.Encode(file, img)
		if err != nil {
			log.Fatal(err)
		}
	}
}
