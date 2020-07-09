package router

import (
	"shila/config"
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
		return trowAwayOddPaths(paths), 0
	}

	expandedSubset := createInitialPathSubsets(paths)
	for expandedSubset.nOfDiffSubsets > 1 && expandedSubset.sizeOfEachSubset < nPathsRequested {
		expandedSubset = expandSubset(paths, expandedSubset)
	}

	return pickSharabilityOptSubset(paths, expandedSubset)
}

func pickSharabilityOptSubset(paths []PathWrapper, subsets pathSubsets) ([]PathWrapper, int) {

	sort.Slice(subsets.subsets, func(i, j int) bool {
		return subsets.subsets[i].sharability < subsets.subsets[j].sharability
	})

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

func expandSubset(paths []PathWrapper, currentSubsets pathSubsets) (subsets pathSubsets) {

	// Selection of the sharability optimal path subset is done greedy. So its not the real optimum.

	nPathsAvailable := len(paths)
	expandedSubsets := make([]pathSubset, 0)

	sort.Slice(currentSubsets.subsets, func(i, j int) bool {
		return currentSubsets.subsets[i].sharability < currentSubsets.subsets[j].sharability
	})

	bestSubsetGreedy := currentSubsets.subsets[0]

	for newIndex := 0; newIndex < nPathsAvailable; newIndex++ {
		expandedSubsets = append(expandedSubsets, pathSubset{
			pathIndices: append(bestSubsetGreedy.pathIndices, newIndex),
			edgeIndices: append(bestSubsetGreedy.edgeIndices, paths[newIndex].edgeIndices...),
		})
	}

	subsets = pathSubsets{
		subsets:          expandedSubsets,
		nOfDiffSubsets:   len(expandedSubsets),
		sizeOfEachSubset: len(expandedSubsets[0].pathIndices),
	}

	calculateSharabilityForPathSubsets(subsets)
	return
}

func createInitialPathSubsets(paths []PathWrapper) (subsets pathSubsets)  {

	setEdgeIndices(paths)

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
	subsets = pathSubsets{subsets: initialSubsets, sizeOfEachSubset: 2, nOfDiffSubsets: len(initialSubsets)}

	calculateSharabilityForPathSubsets(subsets)
	return
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