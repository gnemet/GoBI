package config

import (
	"os"

	"github.com/go-yaml/yaml"
)

type Repository struct {
	Meta    Meta     `yaml:"repository"`
	Reports []Report `yaml:"reports"`
}

type Meta struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type Report struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	TableName    string   `yaml:"table_name"`
	Schema       string   `yaml:"schema"`
	ViewType     string   `yaml:"view_type"`
	ParentReport string   `yaml:"parent_report"`
	ParentColumn string   `yaml:"parent_column"`
	Columns      []Column `yaml:"columns"`
}

type Column struct {
	Name          string `yaml:"name"`
	Label         string `yaml:"label"`
	Type          string `yaml:"type"`
	Filterable    bool   `yaml:"filterable"`
	Sortable      bool   `yaml:"sortable"`
	AggregateFunc string `yaml:"aggregate_func"`
	Hidden        bool   `yaml:"hidden"`
}

func LoadRepository(path string) (*Repository, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var repo Repository
	err = yaml.Unmarshal(data, &repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}
