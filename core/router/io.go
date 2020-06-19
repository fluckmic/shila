//
package router

import (
	"encoding/json"
	"io/ioutil"
	"shila/config"
	"shila/core/shila"
	"shila/io/structure"
	"shila/log"
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

func (router *Router) batchInsert(entries []structure.RoutingEntryJSON) error {

	// Invalid entries are silently ignored and not inserted!
	for _, entry := range entries {

		ipAddressPort, err := entry.Key.GetIPAddressPort()
		if err != nil {
			log.Error.Println(router.Says(PrependError(err, "Skipped insertion of routing Entry.").Error()))
			continue
		}
		key := shila.GetIPAddressPortKey(ipAddressPort)

		dst, err := entry.Flow.GetNetworkAddress()
		if err != nil {
			log.Error.Println(router.Says(PrependError(err, "Skipped insertion of routing Entry.").Error()))
			continue
		}

		if err := router.InsertDestinationFromIPAddressPortKey(key, dst); err != nil {
			log.Error.Println(router.Says(PrependError(err, "Skipped insertion of routing Entry.").Error()))
			continue
		}

		//log.Verbose.Println(router.Says(fmt.Sprint("Inserted routing Entry {", entry, "}.")))
	}

	return nil
}

