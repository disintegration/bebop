# letteravatar
[![GoDoc](https://godoc.org/github.com/disintegration/letteravatar?status.svg)](https://godoc.org/github.com/disintegration/letteravatar)
[![Build Status](https://travis-ci.org/disintegration/letteravatar.svg?branch=master)](https://travis-ci.org/disintegration/letteravatar)

Letter avatar generation for Go.

## Usage

Generate a 100x100px 'A'-letter avatar:

```go
img, err := letteravatar.Draw(100, 'A', nil)
```

The third parameter `options *Options` can be used for customization:

```go
type Options struct {
	Font        *truetype.Font
	Palette     []color.Color
	LetterColor color.Color
	PaletteKey  string
}
```

Using a custom palette:

```go
img, err := letteravatar.Draw(100, 'A', &letteravatar.Options{
	Palette: []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
	},
})
```

## Documentation

[https://godoc.org/github.com/disintegration/letteravatar](https://godoc.org/github.com/disintegration/letteravatar)

## Examples

![](example/Alice.png)
![](example/Bob.png)
![](example/Carol.png)
![](example/Dave.png)
![](example/Eve.png)
![](example/Frank.png)
![](example/Gloria.png)
![](example/Henry.png)
![](example/Isabella.png)
![](example/Жозефина.png)
![](example/Ярослав.png)

```go
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

```

## License

The package "letteravatar" is distributed under the terms of the MIT license.

The Roboto-Medium font is distributed under the terms of the Apache License v2.0.

See [LICENSE](LICENSE).