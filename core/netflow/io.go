//
package netflow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"shila/config"
	"shila/core/shila"
	"shila/io/structure"
	"shila/log"
	"shila/networkSide/network"
)

func loadRoutingEntriesFromDisk() ([]structure.RoutingEntryJSON, error) {

	data, err := ioutil.ReadFile(config.Config.NetFlow.Path)
	if err != nil {
		return nil, err
	}

	var entries []structure.RoutingEntryJSON
	err = json.Unmarshal(data, &entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (r *Router) batchInsert(entries []structure.RoutingEntryJSON) error {

	// Invalid entries are silently ignored and not inserted!
	for _, entry := range entries {

		ipAddressPort, err := entry.Key.GetIPAddressPort()
		if err != nil {
			log.Error.Println(r.Says(PrependError(err, "Skipped insertion of routing entry.").Error()))
			continue
		}
		key := shila.GetIPAddressPortKey(ipAddressPort)

		dst, path, err := entry.Flow.GetNetworkAddressAndPath()
		if err != nil {
			log.Error.Println(r.Says(PrependError(err, "Skipped insertion of routing entry.").Error()))
			continue
		}

		flow := shila.NetFlow{
			Src:  network.AddressGenerator{}.NewEmpty(),
			Path: path,
			Dst:  dst,
		}

		if err := r.InsertFromIPAddressPortKey(key, flow); err != nil {
			log.Error.Println(r.Says(PrependError(err, "Skipped insertion of routing entry.").Error()))
			continue
		}

		log.Verbose.Println(r.Says(fmt.Sprint("Inserted routing entry {", entry, "}.")))
	}

	return nil
}

