// TODO: Add a description.
package config

import "fmt"

var _ config = (*Config)(nil)

type Error string

func (e Error) Error() string {
	return string(e)
}

type config interface {
	InitDefault() error
}

type Config struct {
	KernelEndpoint 	KernelEndpoint
	NetworkSide    	NetworkSide
	NetworkEndpoint NetworkEndpoint
	WorkingSide	   	WorkingSide
}

func (c *Config) InitDefault() (err error) {

	defer func() {
		if e := recover(); e != nil {
			err = e.(Error)
		}
	}()

	// Initialize configuration for the kernel endpoint
	if err = c.KernelEndpoint.InitDefault(); err != nil {
		return Error(fmt.Sprint("Unable to initialize default "+
			"config for kernel endpoint - ", err.Error()))
	}

	// Initialize configuration for the network side
	if err = c.NetworkSide.InitDefault(); err != nil {
		return Error(fmt.Sprint("Unable to initialize default "+
			"config for network side - ", err.Error()))
	}

	// Initialize configuration for the network endpoint
	if err = c.NetworkEndpoint.InitDefault(); err != nil {
		return Error(fmt.Sprint("Unable to initialize default "+
			"config for network endpoint - ", err.Error()))
	}

	// Initialize configuration for the working side
	if err = c.WorkingSide.InitDefault(); err != nil {
		return Error(fmt.Sprint("Unable to initialize default "+
			"config for working side - ", err.Error()))
	}

	return nil
}
