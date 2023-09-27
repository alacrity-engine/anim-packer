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

// SpritesheetMeta is spritesheet metadata
// read from the YAML file.
type SpritesheetMeta struct {
	Name   string `yaml:"name"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}
