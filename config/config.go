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
	Logging        Logging
	KernelEndpoint KernelEndpoint
}

func (c *Config) InitDefault() (err error) {

	defer func() {
		if e := recover(); e != nil {
			err = e.(Error)
		}
	}()

	// Initialize configuration for the logging
	if err = c.Logging.InitDefault(); err != nil {
		return Error(fmt.Sprint("Unable to initialize default "+
			"config for logging - ", err.Error()))
	}

	// Initialize configuration for the application endpoint
	if err = c.KernelEndpoint.InitDefault(); err != nil {
		return Error(fmt.Sprint("Unable to initialize default "+
			"config for application endpoint - ", err.Error()))
	}

	return nil
}
