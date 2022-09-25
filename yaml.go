package main

import (
	"gopkg.in/yaml.v2"
)

// ReadAnimationsData deserializes the animations
// metadata stored as YAML.
func ReadAnimationsData(data []byte) ([]AnimationMeta, error) {
	animations := make([]AnimationMeta, 0)
	err := yaml.Unmarshal(data, &animations)

	if err != nil {
		return nil, err
	}

	return animations, nil
}

// ReadSpritesheetsData deserializes the spritesheets
// metadata stored as YAML.
func ReadSpritesheetsData(data []byte) (map[string]SpritesheetMeta, error) {
	spritesheets := make(map[string]SpritesheetMeta, 0)
	err := yaml.Unmarshal(data, &spritesheets)

	if err != nil {
		return nil, err
	}

	return spritesheets, nil
}
