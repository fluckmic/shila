// TODO: Add a description.
package config

type Error string
func (e Error) Error() string {
	return string(e)
}

var _ config = (*Config) (nil)

type config interface {
	InitDefault() error
}

type Config struct {
	Logging Logging
}

func (c *Config) InitDefault() (err error) {

	defer func() {
		if e := recover(); e != nil {
			err = e.(Error)
		}
	}()

	// Initialize configuration for the logging
	c.Logging.InitDefault()

	return nil
}
