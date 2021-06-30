package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"os"
	"path"
	"strings"

	codec "github.com/alacrity-engine/resource-codec"
	"github.com/boltdb/bolt"
	"github.com/faiface/pixel"
)

var (
	spritesheetsPath         string
	animationsIndexPath      string
	spritesheetsMetadataPath string
	resourceFilePath         string
)

func parseFlags() {
	flag.StringVar(&spritesheetsPath, "spritesheets", "./spritesheets",
		"Path to the directory where spritesheets are stored.")
	flag.StringVar(&animationsIndexPath, "animations-meta", "./animations-meta.yml",
		"Path to the file where animation descriptions are stored.")
	flag.StringVar(&spritesheetsMetadataPath, "spritesheets-meta",
		"./spritesheets-meta.yml", "Path to the spritesheets metadata file.")
	flag.StringVar(&resourceFilePath, "out", "./stage.res",
		"Resource file to store animations and spritesheets.")

	flag.Parse()
}

func loadPicture(pic string) (*pixel.PictureData, error) {
	file, err := os.Open(pic)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	return pixel.PictureDataFromImage(img), nil
}

func main() {
	parseFlags()

	// Get spritesheets from the directory.
	spritesheets, err := ioutil.ReadDir(spritesheetsPath)
	handleError(err)
	// Open the resource file.
	resourceFile, err := bolt.Open(resourceFilePath, 0666, nil)
	handleError(err)
	defer resourceFile.Close()

	// Create collections for spritesheets, animations and tags.
	err = resourceFile.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("spritesheets"))

		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("animations"))

		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("tags"))

		if err != nil {
			return err
		}

		return nil
	})
	handleError(err)

	for _, spritesheetInfo := range spritesheets {
		if spritesheetInfo.IsDir() {
			fmt.Println("Error: directory found in the spritesheets folder.")
			os.Exit(1)
		}

		// Load the spritesheet picture.
		spritesheet, err := loadPicture(path.Join(spritesheetsPath,
			spritesheetInfo.Name()))
		handleError(err)

		// Serialize picture data to byte array.
		spritesheetBytes, err := codec.PictureDataToBytes(spritesheet)
		handleError(err)

		// Save the spritesheet to the database.
		spritesheetID := strings.TrimSuffix(path.Base(spritesheetInfo.Name()),
			path.Ext(spritesheetInfo.Name()))
		err = resourceFile.Update(func(tx *bolt.Tx) error {
			buck := tx.Bucket([]byte("spritesheets"))

			if buck == nil {
				return fmt.Errorf("no spritesheets bucket present")
			}

			err = buck.Put([]byte(spritesheetID), spritesheetBytes)

			if err != nil {
				return err
			}

			return nil
		})
		handleError(err)
	}

	// Read animations data.
	contents, err := ioutil.ReadFile(animationsIndexPath)
	handleError(err)
	animationsMeta, err := codec.ReadAnimationsData(contents)
	handleError(err)

	// Read animation tags.
	animTags := map[string][]string{}

	for _, animMeta := range animationsMeta {
		tag := animMeta.Tag

		// If the tag is absent - create it.
		if _, ok := animTags[tag]; !ok {
			animTags[tag] = []string{}
		}

		// Add the animation name to the tag.
		animTags[tag] = append(animTags[tag],
			animMeta.Name)
	}

	// Read spritesheets data.
	contents, err = ioutil.ReadFile(spritesheetsMetadataPath)
	handleError(err)
	spritesheetsMeta, err := codec.ReadSpritesheetsData(contents)
	handleError(err)

	for _, animMeta := range animationsMeta {
		// Load the animation's spritesheet.
		var picData []byte

		err = resourceFile.View(func(tx *bolt.Tx) error {
			buck := tx.Bucket([]byte("spritesheets"))

			if buck == nil {
				return fmt.Errorf("no spritesheets bucket present")
			}

			picData = buck.Get([]byte(animMeta.Spritesheet))

			if picData == nil {
				return fmt.Errorf("no spritesheet named '%s' found",
					animMeta.Spritesheet)
			}

			return nil
		})
		handleError(err)
		picture, err := codec.PictureDataFromBytes(picData)
		handleError(err)
		// Load its metadata.
		spritesheetMeta := spritesheetsMeta[animMeta.Spritesheet]
		// Get animation frames.
		frames := codec.GetSpritesheetFrames(picture, spritesheetMeta.Width, spritesheetMeta.Height)

		// Assemble the animation.
		anim := &codec.AnimationData{
			Spritesheet: animMeta.Spritesheet,
			Frames:      make([]pixel.Rect, 0),
			Durations:   make([]int32, 0),
		}

		for _, frameMeta := range animMeta.Frames {
			anim.Frames = append(anim.Frames, frames[frameMeta[0]])
			anim.Durations = append(anim.Durations, int32(frameMeta[1]))
		}

		// Store the animation in the database.
		animData, err := codec.AnimationDataToBytes(anim)
		handleError(err)
		err = resourceFile.Update(func(tx *bolt.Tx) error {
			buck := tx.Bucket([]byte("animations"))

			if buck == nil {
				return fmt.Errorf("no animations bucket present")
			}

			err = buck.Put([]byte(animMeta.Name), animData)

			if err != nil {
				return err
			}

			return nil
		})
		handleError(err)
	}

	for tagID, tag := range animTags {
		err = resourceFile.Update(func(tx *bolt.Tx) error {
			buck := tx.Bucket([]byte("tags"))

			if buck == nil {
				return fmt.Errorf("no tags bucket present")
			}

			tagData, err := codec.EncodeTag(tag)

			if err != nil {
				return err
			}

			err = buck.Put([]byte(tagID), tagData)

			if err != nil {
				return err
			}

			return nil
		})
	}
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
