package config

var _ config = (*WorkingSide)(nil)

type WorkingSide struct {

	NumberOfWorkerPerChannel int

}

func (ws *WorkingSide) InitDefault() error {

	ws.NumberOfWorkerPerChannel = 1

	return nil
}
