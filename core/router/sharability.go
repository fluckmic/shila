package router

import (
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

func getSharabilityOptSubset(paths []PathWrapper) ([]PathWrapper, int) {

	// If there is just one path available, (or none), we cannot choose..
	if len(paths) < 2 {
		return paths, 0
	}

	nPathsRequested := config.Config.KernelSide.NumberOfEgressInterfaces

	// If there is one path requested, every path is optimal w.r.t. sharability
	if nPathsRequested < 2 {
		return paths, 0
	}

	setEdgeIndices(paths)

	mergedSubsets 	 := createInitialPathSubsets(paths)
	for mergedSubsets.nOfDiffSubsets > 1 && mergedSubsets.sizeOfEachSubset < nPathsRequested {
		mergedSubsets = mergeSubsets(paths, mergedSubsets)
	}

	calculateSharabilityForPathSubsets(mergedSubsets)

	return pickSharabilityOptSubset(paths, mergedSubsets)
}

func pickSharabilityOptSubset(paths []PathWrapper, subsets pathSubsets) ([]PathWrapper, int) {

	sort.Slice(subsets.subsets, func(i, j int) bool {
		return subsets.subsets[i].sharability < subsets.subsets[j].sharability
	})

	log.Info.Print("Number of sharability optimal subsets to choose from: ", subsets.nOfDiffSubsets, ".")

	sharabilityOptSubset := make([]PathWrapper, 0)
	for _, pathIndex := range subsets.subsets[0].pathIndices {
		sharabilityOptSubset = append(sharabilityOptSubset, paths[pathIndex])
	}

	return sharabilityOptSubset, subsets.subsets[0].sharability
}

func calculateSharabilityForPathSubsets(subsets pathSubsets) {
	for indexCurrentSubset, currentSubset := range subsets.subsets {
		sort.Ints(currentSubset.edgeIndices)
		for i := 0; i < len(currentSubset.edgeIndices)-1; i++ {
			if currentSubset.edgeIndices[i] == currentSubset.edgeIndices[i+1] {
				subsets.subsets[indexCurrentSubset].sharability++
			}
		}
	}
}

func calculateSharabilityForPaths(paths []PathWrapper) (sharabilityValue int) {

	// First determine the edge indices
	setEdgeIndices(paths)

	// Merge all edge indices
	allEdgeIndices := make([]int, 0)
	for _, path := range paths {
		allEdgeIndices = append(allEdgeIndices, path.edgeIndices...)
	}

	// Sort them
	sort.Ints(allEdgeIndices)

	// Calculate the sharability
	for i := 0; i < len(allEdgeIndices)-1; i++ {
		if allEdgeIndices[i] == allEdgeIndices[i+1] {
			sharabilityValue++
		}
	}
	return
}

func mergeSubsets(paths []PathWrapper, currentSubsets pathSubsets) pathSubsets {

	nPathsAvailable := len(paths)
	mergedSubsets 	:= make([]pathSubset, 0)

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

func createInitialPathSubsets(paths []PathWrapper) pathSubsets {
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