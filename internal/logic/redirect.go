package logic

import (
	"fmt"
	"log"
	"strings"

	"github.com/tanay13/GlitchMesh/internal/client"
	"github.com/tanay13/GlitchMesh/internal/config"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

func Redirect(path string) ([]byte, error) {
	yaml_file_path := config.Configs.Env.YAML_FILE_PATH

	if yaml_file_path == "" {
		log.Fatalf("Please set up the proxy yaml first!!")
	}

	proxy_config, err := utils.ParseYaml(yaml_file_path)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	urlParts := strings.SplitN(path, "/", 2)

	service_name := urlParts[0]
	endpoint := urlParts[1]

	service_config := utils.GetServiceConfig(service_name, proxy_config)

	if service_config == nil {
		return nil, fmt.Errorf("Service config not found in the proxy list")
	}

	// fault_name := service_config.Fault
	service_url := service_config.Url
	//	fault_value := service_config.Value

	completeUrl := service_url + endpoint

	resp, err := client.MakeGetRedirection(completeUrl)

	if err != nil {

		return nil, fmt.Errorf("Error during service call %v", err)

	}

	return resp, nil
}
