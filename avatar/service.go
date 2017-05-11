// Package avatar provides a service that manages bebop user avatars.
package avatar

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"unicode"
	"unicode/utf8"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/disintegration/gift"
	"github.com/disintegration/letteravatar"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/satori/go.uuid"

	"github.com/disintegration/bebop/filestorage"
	"github.com/disintegration/bebop/store"
)

const (
	avatarSize   = 100
	minImageSize = 50
	maxImageSize = 2000
)

// Input image format/size errors.
var (
	ErrImageDecode   = errors.New("avatar: image decode failed")
	ErrImageTooSmall = errors.New("avatar: image too small")
	ErrImageTooLarge = errors.New("avatar: image too large")
)

// Service is an avatar-processing service.
type Service interface {
	// Save preprocesses and saves a new avatar for the given user.
	Save(user *store.User, imageData []byte) error

	// Generate generates a new avatar for the given user.
	Generate(user *store.User) error

	// URL returns the avatar URL of the given user.
	URL(user *store.User) string
}

// service is the main implementation of the Service.
type service struct {
	fileStorage filestorage.FileStorage
	userStore   store.UserStore
	logger      *log.Logger
}

// NewService creates a new avatar service.
func NewService(userStore store.UserStore, fileStorage filestorage.FileStorage, logger *log.Logger) Service {
	return &service{
		fileStorage: fileStorage,
		userStore:   userStore,
		logger:      logger,
	}
}

// Save preprocesses and saves a new avatar for the given user.
func (s *service) Save(user *store.User, imageData []byte) error {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return ErrImageDecode
	}

	if cfg.Width < minImageSize || cfg.Height < minImageSize {
		return ErrImageTooSmall
	}
	if cfg.Width > maxImageSize || cfg.Height > maxImageSize {
		return ErrImageTooLarge
	}

	var avatar string
	if format == "gif" {
		avatar, err = s.prepareAndSaveGIF(imageData)
		if err != nil {
			return fmt.Errorf("avatar: prepare and save gif failed: %s", err)
		}
	} else {
		avatar, err = s.prepareAndSaveImage(imageData)
		if err != nil {
			return fmt.Errorf("avatar: prepare and save failed: %s", err)
		}
	}

	err = s.userStore.SetAvatar(user.ID, avatar)
	if err != nil {
		return fmt.Errorf("avatar: set user avatar failed: %s", err)
	}

	// Remove the old avatar file.
	if user.Avatar != "" {
		s.remove(user.Avatar)
	}

	return nil
}

// Generate generates a new avatar for the given user.
func (s *service) Generate(user *store.User) error {
	var letter rune
	if user.Name == "" {
		letter = ' '
	} else {
		letter, _ = utf8.DecodeRuneInString(user.Name)
	}

	img, err := letteravatar.Draw(avatarSize, unicode.ToUpper(letter), nil)
	if err != nil {
		return fmt.Errorf("avatar: letteravatar draw failed: %s", err)
	}

	avatar, err := s.saveImage(img)
	if err != nil {
		return fmt.Errorf("avatar: save avatar file failed: %s", err)
	}

	err = s.userStore.SetAvatar(user.ID, avatar)
	if err != nil {
		return fmt.Errorf("avatar: set user avatar failed: %s", err)
	}

	// Remove the old avatar file.
	if user.Avatar != "" {
		s.remove(user.Avatar)
	}

	return nil
}

// URL returns the avatar URL of the given user.
func (s *service) URL(user *store.User) string {
	if user.Avatar == "" {
		return ""
	}
	return s.fileStorage.URL("avatars/" + user.Avatar)
}

// prepareAndSaveImage preprocesses and saves the given avatar image to the file storage.
func (s *service) prepareAndSaveImage(imageData []byte) (string, error) {
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("avatar: decode avatar image failed: %s", err)
	}

	g := gift.New()
	g.Add(gift.ResizeToFill(avatarSize, avatarSize, gift.CubicResampling, gift.CenterAnchor))

	if format == "jpeg" || format == "tiff" {
		o := readOrientation(bytes.NewReader(imageData))
		if filter, ok := orientationFilters[o]; ok {
			g.Add(filter)
		}
	}

	newImg := image.NewNRGBA(g.Bounds(img.Bounds()))
	g.Draw(newImg, img)

	avatar, err := s.saveImage(newImg)
	if err != nil {
		return "", fmt.Errorf("avatar: save avatar file failed: %s", err)
	}

	return avatar, nil
}

// prepareAndSaveGIF preprocesses and saves the given GIF
// avatar image (possibly animated) to the file storage.
func (s *service) prepareAndSaveGIF(imageData []byte) (string, error) {
	gifImg, err := gif.DecodeAll(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("avatar: decode gif avatar image failed: %s", err)
	}

	if gifImg.Config.Width != avatarSize || gifImg.Config.Height != avatarSize {
		// This resizing method uses the nearest-neighbor resampling (no antialiasing).
		// It does not restore the full image for each frame. The benefit is that
		// the original palette can be used for each image. Also it seems to work fine
		// for animated gifs with transparent areas. Needs more testing.
		copier := gift.New()
		resizer := gift.New(gift.ResizeToFill(avatarSize, avatarSize, gift.NearestNeighborResampling, gift.CenterAnchor))
		frame := image.NewRGBA(image.Rect(0, 0, gifImg.Config.Width, gifImg.Config.Height))
		for i, img := range gifImg.Image {
			if i > 0 {
				for j := range frame.Pix {
					frame.Pix[j] = 0
				}
			}
			copier.DrawAt(frame, img, img.Bounds().Min, gift.CopyOperator)
			dstFrame := image.NewPaletted(image.Rect(0, 0, avatarSize, avatarSize), img.Palette)
			resizer.Draw(dstFrame, frame)
			gifImg.Image[i] = dstFrame
		}
		gifImg.Config.Width = avatarSize
		gifImg.Config.Height = avatarSize
	}

	avatar, err := s.saveGIF(gifImg)
	if err != nil {
		return "", fmt.Errorf("avatar: save avatar file failed: %s", err)
	}

	return avatar, nil
}

// saveImage saves the given image to the file storage.
// The storage format (JPEG or PNG) is determined based on the image opacity.
func (s *service) saveImage(img image.Image) (string, error) {
	opaque := false
	if img, ok := img.(interface {
		Opaque() bool
	}); ok && img.Opaque() {
		opaque = true
	}

	filename := genUniqueFilename()
	if opaque {
		filename += ".jpg"
	} else {
		filename += ".png"
	}

	r, w := io.Pipe()
	defer r.Close()

	go func() {
		if opaque {
			w.CloseWithError(jpeg.Encode(w, img, &jpeg.Options{Quality: 90}))
		} else {
			w.CloseWithError(png.Encode(w, img))
		}
	}()

	err := s.fileStorage.Save("avatars/"+filename, r)
	if err != nil {
		return "", fmt.Errorf("image save failed: %s", err)
	}

	return filename, nil
}

// saveGIF saves the given GIF image (possibly animated) to the file storage.
func (s *service) saveGIF(gifImg *gif.GIF) (string, error) {
	filename := genUniqueFilename()
	filename += ".gif"

	r, w := io.Pipe()
	defer r.Close()

	go func() {
		w.CloseWithError(gif.EncodeAll(w, gifImg))
	}()

	err := s.fileStorage.Save("avatars/"+filename, r)
	if err != nil {
		return "", fmt.Errorf("image save failed: %s", err)
	}

	return filename, nil
}

// remove removes the given avatar from the file storage.
func (s *service) remove(filename string) {
	err := s.fileStorage.Remove("avatars/" + filename)
	if err != nil {
		s.logger.Printf("avatar: failed to remove avatar file: %s", err)
	}
}

func genUniqueFilename() string {
	return uuid.NewV4().String()
}

// readOrientation reads the EXIF orientation tag from the given image.
// It returns 0 if the orientation tag is not found or invalid.
func readOrientation(r io.Reader) int {
	x, err := exif.Decode(r)
	if err != nil {
		return 0
	}

	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 0
	}

	orientation, err := tag.Int(0)
	if err != nil {
		return 0
	}

	if orientation < 1 || orientation > 8 {
		return 0
	}

	return orientation
}

// Filters needed to fix the given image orientation.
var orientationFilters = map[int]gift.Filter{
	2: gift.FlipHorizontal(),
	3: gift.Rotate180(),
	4: gift.FlipVertical(),
	5: gift.Transpose(),
	6: gift.Rotate270(),
	7: gift.Transverse(),
	8: gift.Rotate90(),
}
