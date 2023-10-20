package main

import (
	"flag"
	"fmt"
	_ "image/png"
	"os"
	"path"
	"strings"

	"github.com/alacrity-engine/core/math/geometry"
	codec "github.com/alacrity-engine/resource-codec"
	"github.com/golang-collections/collections/queue"
	bolt "go.etcd.io/bbolt"
)

var (
	projectPath      string
	resourceFilePath string
)

func parseFlags() {
	flag.StringVar(&projectPath, "project", ".",
		"Path to the project to pack animations for.")
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
	animTags := map[string][]string{}

	entries, err := os.ReadDir(projectPath)
	handleError(err)

	traverseQueue := queue.New()

	if len(entries) <= 0 {
		return
	}

	for _, entry := range entries {
		traverseQueue.Enqueue(FileTracker{
			EntryPath: projectPath,
			Entry:     entry,
		})
	}

	for traverseQueue.Len() > 0 {
		fsEntry := traverseQueue.Dequeue().(FileTracker)

		if fsEntry.Entry.IsDir() {
			entries, err = os.ReadDir(path.Join(fsEntry.EntryPath, fsEntry.Entry.Name()))
			handleError(err)

			for _, entry := range entries {
				traverseQueue.Enqueue(FileTracker{
					EntryPath: path.Join(fsEntry.EntryPath, fsEntry.Entry.Name()),
					Entry:     entry,
				})
			}

			continue
		}

		if !strings.HasSuffix(fsEntry.Entry.Name(), ".anim.yml") {
			continue
		}

		// Read animations data.
		contents, err := os.ReadFile(path.Join(fsEntry.EntryPath, fsEntry.Entry.Name()))
		handleError(err)
		animationsMeta, err := ReadAnimationsData(contents)
		handleError(err)

		// Read animation tags.
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

				frames, err := compressedPic.GetSpritesheetFrames(ss)

				if err != nil {
					return err
				}

				// Assemble the animation.
				anim := &codec.AnimationData{
					SpritesheetID: animationMeta.SpritesheetID,
					TextureID:     animationMeta.TextureID,
					Frames:        make([]geometry.Rect, 0),
					Durations:     make([]int32, 0),
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
			handleError(err)
		}
	}

	for tagID, tag := range animTags {
		err = resourceFile.Update(func(tx *bolt.Tx) error {
			buck, err := tx.CreateBucketIfNotExists([]byte("tags"))

			if err != nil {
				return err
			}

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
		handleError(err)
	}
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
