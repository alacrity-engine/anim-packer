package main

// AnimationMeta is animation metadata
// read from the YAML file.
type AnimationMeta struct {
	Name        string   `yaml:"name"`
	Tag         string   `yaml:"tag"`
	Spritesheet string   `yaml:"spritesheet"`
	Frames      [][2]int `yaml:"frames"`
}

// SpritesheetMeta is spritesheet metadata
// read from the YAML file.
type SpritesheetMeta struct {
	Width  int
	Height int
}
