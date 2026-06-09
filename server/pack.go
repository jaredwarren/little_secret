package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type WordPair struct {
	Good     string `json:"good"`
	Confused string `json:"confused"`
}

type MissionPack struct {
	Name  string     `json:"name"`
	Words []WordPair `json:"words"`
}

const DefaultPackName = "Classic Pack"

var DefaultWords = []WordPair{
	{"Perfume", "Deodorant"},
	{"Coffee", "Tea"},
	{"Bicycle", "Scooter"},
	{"Ocean", "Lake"},
	{"Apple", "Pear"},
	{"Library", "Bookshop"},
	{"Airplane", "Helicopter"},
	{"Guitar", "Violin"},
	{"Pizza", "Flatbread"},
	{"Cat", "Tiger"},
	{"Sun", "Star"},
	{"River", "Canal"},
	{"Cinema", "Theatre"},
	{"Forest", "Jungle"},
	{"Doctor", "Nurse"},
	{"Watch", "Clock"},
	{"Mountain", "Hill"},
	{"Castle", "Palace"},
	{"Window", "Mirror"},
	{"Bread", "Cake"},
	{"Gold", "Brass"},
}

// LoadPacks loads all mission packs from the specified directory.
// If the directory is empty, it writes the default Classic Pack first.
func LoadPacks(dir string) (map[string]MissionPack, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create pack directory: %w", err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read pack directory: %w", err)
	}

	// If no packs exist, seed the default pack
	if len(files) == 0 {
		defaultPack := MissionPack{
			Name:  DefaultPackName,
			Words: DefaultWords,
		}
		if err := SavePack(dir, defaultPack); err != nil {
			return nil, fmt.Errorf("failed to save default pack: %w", err)
		}
		files, err = os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to re-read pack directory: %w", err)
		}
	}

	packs := make(map[string]MissionPack)
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // skip unreadable files
		}

		var pack MissionPack
		if err := json.Unmarshal(data, &pack); err != nil {
			continue // skip invalid JSON files
		}

		packs[pack.Name] = pack
	}

	return packs, nil
}

// SavePack saves a mission pack to the specified directory.
func SavePack(dir string, pack MissionPack) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create pack directory: %w", err)
	}

	filename := fmt.Sprintf("%s.json", filepath.Base(pack.Name)) // simple sanitization
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pack: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write pack file: %w", err)
	}

	return nil
}
