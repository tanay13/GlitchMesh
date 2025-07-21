package models

type Config struct {
	Env struct {
		YAML_FILE_PATH string `json:"yaml_file_path"`
	} `json:"env"`
}
