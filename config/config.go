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

	if Config.Config.DumpConfig {
		dumpConfig()
	}
}

func dumpConfig() {
	// Dump the applied config file
	if configDump, err := json.Marshal(Config); err != nil {
		fmt.Print(err)
	} else {
		if err := ioutil.WriteFile(Config.Config.ConfigDumpPath, configDump, 0644); err != nil {
			fmt.Print(err)
		}
	}
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
			VacuumInterval:                      		5,
			MaxTimeUntouched:                    		300,
			WaitingTimeTrafficConnEstablishment: 		1,
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
			SizeIngressBuffer:          				250,
			SizeEgressBuffer:           				250,
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
			SizeIngressBuffer:              		 	250,
			SizeEgressBuffer:                		 	250,
			SizeRawIngressStorage:           		 	2500,
			WaitingTimeAfterConnectionIssue: 		 	2,
			ServerResendInterval:            		 	2,
			SizeHoldingArea:                 		 	100,
		},
		Router: structure.RouterConfigJSON{
			PathSelection: 								"shortest",
		},
		Config: structure.ConfigConfigJSON{
			DumpConfig:									false,
			ConfigDumpPath:								"_config.dump",
		},
	}
}
