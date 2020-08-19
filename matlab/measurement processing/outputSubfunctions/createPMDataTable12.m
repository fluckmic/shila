function [PMDataTable1, PMDataTable2] = createPMDataTable12(exp)

    % N. Interf. | MTU GP/TP (avg, std) | Len. GP/TP (avg, std) | Shar. GP/TP (avg, std)
    PMDataTable1 = zeros(exp.nInterfaceCounts, 4 * exp.nPathSelections + 1);
    PMDataTable2 = zeros(max(exp.clients), max(exp.clients), exp.nInterfaceCounts, 4 * exp.nPathSelections + 1);
    
    for pathSelectionIdIndex = 1:numel(exp.pathSelections)

        pathSelectionId = exp.pathSelections(pathSelectionIdIndex);
        index = ((pathSelectionId - 1) * 4) + 2;
        
        goodputPathSelectionCS = [];
        goodputPathSelectionSC = [];
        goodputPathSelection   = [];

        throughputPathSelectionCS = [];
        throughputPathSelectionSC = [];
        throughputPathSelection   = [];

        for clientHostIdIndex = 1:numel(exp.clients)
            for serverHostIdIndex = 1:numel(exp.clients)

                clientHostId = exp.clients(clientHostIdIndex);
                serverHostId = exp.clients(serverHostIdIndex);

                if clientHostId == serverHostId
                    continue
                end

                % Goodput
                % +++++++
                % Get all measurements taken at the server when sending from client to server
                % path selection | client | server | measurement side | sending direction | interface | repetition | time | measurement value
                goodputClientServerPairCS = exp.dataCubusIperf(pathSelectionId, clientHostId, serverHostId, 2, 1, :, :, :, 2);
                goodputClientServerPairCS = reshape(goodputClientServerPairCS, size(goodputClientServerPairCS,6:8));
                goodputClientServerPairCS = mean(goodputClientServerPairCS, 3);
                goodputClientServerPairCS = goodputClientServerPairCS(exp.interfaces, :);
               
                % Get all measurements taken at the client when sending from server to client
                goodputClientServerPairSC = exp.dataCubusIperf(pathSelectionId, clientHostId, serverHostId, 1, 2, :, :, :, 2);
                goodputClientServerPairSC = reshape(goodputClientServerPairSC, size(goodputClientServerPairSC,6:8));
                goodputClientServerPairSC = mean(goodputClientServerPairSC, 3);
                goodputClientServerPairSC = goodputClientServerPairSC(exp.interfaces, :);

                % Combine them
                goodputClientServerPair   = [goodputClientServerPairCS];
                %goodputClientServerPair   = [goodputClientServerPairCS goodputClientServerPairSC];
                
                % Take average and std (over repetitions, per sending direction)
                avgGoodputClientServerPairCS = mean(goodputClientServerPairCS, 2);
                stdGoodputClientServerPairCS = std(goodputClientServerPairCS, 1, 2);

                avgGoodputClientServerPairSC = mean(goodputClientServerPairSC, 2);
                stdGoodputClientServerPairSC = std(goodputClientServerPairSC, 1, 2);

                % Take average and std (over repetitions, for both sending directions)
                avgGoodputClientServerPair = mean(goodputClientServerPair, 2);
                stdGoodputClientServerPair = std(goodputClientServerPair, 1, 2);

                % Concatenate them
                goodputPathSelectionCS = [goodputPathSelectionCS, goodputClientServerPairCS];
                goodputPathSelectionSC = [goodputPathSelectionSC, goodputClientServerPairSC];
                goodputPathSelection   = [goodputPathSelection, goodputClientServerPair];

                % Throughput
                % ++++++++++
                % Get all measurements taken at the server when sending from client to server
                throughputClientServerPairCS = exp.dataCubusTSharkSCION(pathSelectionId, clientHostId, serverHostId, 1, :, :, :);
                throughputClientServerPairCS = reshape(throughputClientServerPairCS, size(throughputClientServerPairCS,5:7));
                throughputClientServerPairCS = throughputClientServerPairCS(exp.interfaces, :);
                % Get all measurements taken at the client when sending from server to client
                throughputClientServerPairSC = exp.dataCubusTSharkSCION(pathSelectionId, clientHostId, serverHostId, 2, :, :, :);
                throughputClientServerPairSC = reshape(throughputClientServerPairSC, size(throughputClientServerPairSC,5:7));
                throughputClientServerPairSC = throughputClientServerPairSC(exp.interfaces, :);

                % Combine them
                throughputClientServerPair = [throughputClientServerPairCS throughputClientServerPairSC];

                % Take average and std (over repetitions, per sending direction)
                avgThroughputClientServerPairCS = mean(throughputClientServerPairCS, 2);
                stdThroughputClientServerPairCS = std(throughputClientServerPairCS, 1, 2);

                avgGThroughputClientServerPairSC = mean(throughputClientServerPairSC, 2);
                stdGThroughputClientServerPairSC = std(throughputClientServerPairSC, 1, 2);

                % Take average and std (over repetitions, for both sending directions)
                avgThroughputClientServerPair = mean(throughputClientServerPair, 2);
                stdThroughputClientServerPair = std(throughputClientServerPair, 1, 2);

                % Concatenate them
                throughputPathSelectionCS = [throughputPathSelectionCS, throughputClientServerPairCS];
                throughputPathSelectionSC = [throughputPathSelectionSC, throughputClientServerPairSC];
                throughputPathSelection   = [throughputPathSelection, throughputClientServerPair];

                % Assign the measures
                PMDataTable2(clientHostId, serverHostId, :, index:(index + 3)) = [avgGoodputClientServerPair, stdGoodputClientServerPair, avgThroughputClientServerPair, stdThroughputClientServerPair];
                PMDataTable2(clientHostId, serverHostId, :,1) = exp.interfaces;
                
                
            end
        end

        % Averaging and stding the goodput
        AvgGoodputPathSelectionCS = mean(goodputPathSelectionCS,2);
        AvgGoodputPathSelectionSC = mean(goodputPathSelectionSC,2);
        AvgGoodputPathSelection   = mean(goodputPathSelection,2);

        StdGoodputPathSelectionCS = std(goodputPathSelectionCS,1,2);
        StdGoodputPathSelectionSC = std(goodputPathSelectionSC,1,2);
        StdGoodputPathSelection   = std(goodputPathSelection,1,2);

        % Averaging and stding the throughput
        AvgThroughputPathSelectionCS = mean(throughputPathSelectionCS,2);
        AvgThroughputPathSelectionSC = mean(throughputPathSelectionSC,2);
        AvgThroughputPathSelection   = mean(throughputPathSelection,2);

        StdThroughputPathSelectionCS = std(throughputPathSelectionCS,1,2);
        StdThroughputPathSelectionSC = std(throughputPathSelectionSC,1,2);
        StdThroughputPathSelection   = std(throughputPathSelection,1,2);

        % Assign the measures
        PMDataTable1(:, index:(index + 3)) = [AvgGoodputPathSelection, StdGoodputPathSelection, AvgThroughputPathSelection, StdThroughputPathSelection];
        PMDataTable1(:,1) = exp.interfaces;
        
    end
end

