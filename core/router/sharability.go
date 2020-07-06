package router

import (
	"fmt"
	"math/big"
	"shila/config"
	"shila/log"
	"sort"
)

type edge string

type pathSubset struct {
	pathIndices []int
	edgeIndices []int
	sharability int
}

type pathSubsets struct {
	subsets          []pathSubset
	sizeOfEachSubset int
	nOfDiffSubsets   int
}

func getSharabilityOptSubset(paths []PathWrapper) []PathWrapper {

	nPathsAvailable := len(paths)
	nPathsRequested := config.Config.KernelSide.NumberOfEgressInterfaces

	log.Info.Print("Determine sharability optimum pathSubset for ", nPathsAvailable, " paths.\n")
	for _, path := range paths {
		fmt.Printf("%v\n", path.path)
	}
	log.Info.Print(big.NewInt(0).Binomial(int64(nPathsAvailable), int64(nPathsAvailable)), " possibilities for ",
		nPathsRequested, " different paths taken.\n")

	setEdgeIndices(paths)

	mergedSubsets 	 := createInitialSubsets(paths)
	for mergedSubsets.nOfDiffSubsets > 1 && mergedSubsets.sizeOfEachSubset < nPathsRequested {
		mergedSubsets = mergeSubsets(nPathsAvailable, paths, mergedSubsets)
	}

	calculateSharability(mergedSubsets)

	return nil
}

func calculateSharability(subsets pathSubsets) {
	for indexCurrentSubset, currentSubset := range subsets.subsets {
		sort.Ints(currentSubset.edgeIndices)
		for i := 0; i < len(currentSubset.edgeIndices)-1; i++ {
			if currentSubset.edgeIndices[i] == currentSubset.edgeIndices[i+1] {
				subsets.subsets[indexCurrentSubset].sharability++
			}
		}
	}
}

func mergeSubsets(nPathsAvailable int, paths []PathWrapper, currentSubsets pathSubsets) pathSubsets {

	mergedSubsets := make([]pathSubset, 0)

	for _, currentSubset := range currentSubsets.subsets {

		highestIndexSoFar := currentSubset.pathIndices[len(currentSubset.pathIndices)-1]
		newIndex		  := highestIndexSoFar + 1

		if nPathsAvailable <= newIndex {
			continue
		}

		mergedSubsets = append(mergedSubsets, pathSubset{
			pathIndices: append(currentSubset.pathIndices, newIndex),
			edgeIndices: append(currentSubset.edgeIndices, paths[newIndex].edgeIndices...),
		})
	}

	return pathSubsets{
		subsets:          mergedSubsets,
		nOfDiffSubsets:   len(mergedSubsets),
		sizeOfEachSubset: len(mergedSubsets[0].pathIndices),
	}
}

func createInitialSubsets(paths []PathWrapper) pathSubsets {
	nPaths := len(paths)
	initialSubsets := make([]pathSubset,0)
	for i := 0; i < nPaths; i++ {
		for j := i+1; j < nPaths; j++ {
			initialSubsets = append(initialSubsets, pathSubset{
				pathIndices: []int{i,j},
				edgeIndices: append(paths[i].edgeIndices, paths[j].edgeIndices...),
			})
		}
	}
	return pathSubsets{subsets: initialSubsets, sizeOfEachSubset: 2, nOfDiffSubsets: len(initialSubsets)}
}

// Returns the number of distinct edges in paths
func setEdgeIndices(paths []PathWrapper) {

	edgeToIndex 	:= make(map [edge] int)	// Map an edge to an index.

	nextEdgeIndex := 0
	for pathIndex, path := range paths {
		paths[pathIndex].edgeIndices = make([]int, 0)	// Indices of all edges of a certain path
		for i := 0; i < len(path.path.Interfaces())-1; i++ {

			edge := edge(	path.path.Interfaces()[i].ID().String() 	+ "-" +
							path.path.Interfaces()[i].IA().String() 	+ "-" +
							path.path.Interfaces()[i+1].ID().String() 	+ "-" +
							path.path.Interfaces()[i+1].IA().String())

			// new, unknown edge
			if _, ok := edgeToIndex[edge]; !ok {
				edgeToIndex[edge] = nextEdgeIndex
				paths[pathIndex].edgeIndices = append(paths[pathIndex].edgeIndices, nextEdgeIndex)
				nextEdgeIndex++
			// already known edge
			} else {
				paths[pathIndex].edgeIndices = append(paths[pathIndex].edgeIndices, edgeToIndex[edge])
			}
		}
	}
}