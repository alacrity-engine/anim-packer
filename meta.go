package main

// AnimationMeta is animation metadata
// read from the YAML file.
type AnimationMeta struct {
	Name          string   `yaml:"name"`
	Tag           string   `yaml:"tag"`
	TextureID     string   `yaml:"textureID"`
	SpritesheetID string   `yaml:"spritesheetID"`
	Frames        [][2]int `yaml:"frames"`
}
