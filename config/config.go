package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"shila/io/structure"
)

var Config structure.ConfigJSON

func init() {
	Config = loadConfig()
}

func loadConfig() structure.ConfigJSON {

	// Load the default values
	configJSON := defaultConfig()

	// Get the path to the config file from the command line argument.
	configPath := flag.String("config", "", "Path to the config file."); flag.Parse()
	if *configPath == "" {
		return *configJSON
	}

	// Try to parse the config file.
	if err := loadConfigFromDisk(configJSON, *configPath); err != nil {
		fmt.Print(err)
		return *configJSON
	}

	return *configJSON
}

func loadConfigFromDisk(config *structure.ConfigJSON, path string) error {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return nil
}

func defaultConfig() *structure.ConfigJSON {
	return &structure.ConfigJSON{
		WorkingSide:     structure.WorkingSideConfigJSON{
			NumberOfWorkerPerChannel: 			 		1,
		},
		Connection:      structure.ConnectionConfigJSON{
			VacuumInterval:                      		1,
			MaxTimeUntouched:                    		300,
			WaitingTimeTrafficConnEstablishment: 		2,
		},
		NetFlow:         structure.NetFlowConfigJSON{
			Path: 								 		"routing.json",
		},
		KernelSide:      structure.KernelSideConfigJSON{
			NumberOfEgressInterfaces: 					3,
			EgressNamespace:          					"shila-egress",
			IngressNamespace:         					"shila-ingress",
			IngressIP:                					"10.7.0.9",
		},
		KernelEndpoint:  structure.KernelEndpointConfigJSON{
			SizeIngressBuffer:          				100,
			SizeEgressBuffer:           				100,
			SizeRawIngressBuffer:       				500,
			SizeRawIngressStorage:      				2500,
			ReadSizeRawIngress:         				30,
			WaitingTimeUntilEscalation: 				5,
		},
		Logging:         structure.LoggingConfigJSON{
			PrintVerbose: 								false,
		},
		NetworkSide:     structure.NetworkSideConfigJSON{
			ContactingServerPort: 						9876,
		},
		NetworkEndpoint: structure.NetworkEndpointConfigJSON{
			SizeIngressBuffer:              		 	100,
			SizeEgressBuffer:                		 	100,
			SizeRawIngressStorage:           		 	2500,
			WaitingTimeAfterConnectionIssue: 		 	2,
			ServerResendInterval:            		 	2,
			SizeHoldingArea:                 		 	100,
		},
		Router: structure.RouterConfigJSON{
			PathSelection: 								"shortest",
		},
	}
}
