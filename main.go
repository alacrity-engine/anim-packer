package main

import (
	"flag"
	"fmt"
	_ "image/png"
	"os"

	"github.com/alacrity-engine/core/geometry"
	codec "github.com/alacrity-engine/resource-codec"
	bolt "go.etcd.io/bbolt"
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

func main() {
	parseFlags()

	// Open the resource file.
	resourceFile, err := bolt.Open(resourceFilePath, 0666, nil)
	handleError(err)
	defer resourceFile.Close()

	// Read animations data.
	contents, err := os.ReadFile(animationsIndexPath)
	handleError(err)
	animationsMeta, err := ReadAnimationsData(contents)
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

	// Save everything.
	for _, animationMeta := range animationsMeta {
		err = resourceFile.Update(func(tx *bolt.Tx) error {
			buck := tx.Bucket([]byte("spritesheets"))

			if buck == nil {
				return fmt.Errorf("the spritesheets bucket not found")
			}

			ssBytes := buck.Get([]byte(animationMeta.SpritesheetID))

			if ssBytes == nil {
				return fmt.Errorf(
					"spritesheet '%s' not found", animationMeta.SpritesheetID)
			}

			ss, err := codec.SpritesheetDataFromBytes(ssBytes)

			if err != nil {
				return err
			}

			textureBuck := tx.Bucket([]byte("textures"))

			if textureBuck == nil {
				return fmt.Errorf("the textures bucket not found")
			}

			textureBytes := textureBuck.Get([]byte(animationMeta.TextureID))

			if textureBytes == nil {
				return fmt.Errorf(
					"texture '%s' not found", animationMeta.TextureID)
			}

			texture, err := codec.TextureDataFromBytes(textureBytes)

			if err != nil {
				return err
			}

			picBucket := tx.Bucket([]byte("pictures"))

			if picBucket == nil {
				return fmt.Errorf("the pictures bucket not found")
			}

			picBytes := picBucket.Get([]byte(texture.PictureID))

			if picBytes == nil {
				return fmt.Errorf(
					"picture '%s' not found", texture.PictureID)
			}

			compressedPic, err := codec.CompressedPictureFromBytes(picBytes)

			if err != nil {
				return err
			}

			frames := compressedPic.GetSpritesheetFrames(
				int(ss.Width), int(ss.Height))

			// Assemble the animation.
			anim := &codec.AnimationData{
				TextureID: animationMeta.TextureID,
				Frames:    make([]geometry.Rect, 0),
				Durations: make([]int32, 0),
			}

			for _, frameMeta := range animationMeta.Frames {
				anim.Frames = append(anim.Frames, frames[frameMeta[0]])
				anim.Durations = append(anim.Durations, int32(frameMeta[1]))
			}

			data, err := anim.ToBytes()

			if err != nil {
				return err
			}

			animBucket, err := tx.CreateBucketIfNotExists([]byte("animations"))

			if err != nil {
				return err
			}

			err = animBucket.Put([]byte(animationMeta.Name), data)

			if err != nil {
				return err
			}

			return nil
		})
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
