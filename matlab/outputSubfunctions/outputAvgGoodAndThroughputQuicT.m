function outputAvgGoodAndThroughputQuicT(exp, pathToReportFolder, export)

% client | server | sending direction | repetition | measurement value
goodput     = exp.dataCubusQuicT;

nonGood = reshape((goodput == -1), 1, []);
sum(nonGood);

avgGoodputOverRep  = mean(goodput, 4); % mean over repetition
avgGoodputOverRep = reshape(avgGoodputOverRep, size(avgGoodputOverRep, [1 2 3]));

avgGoodputOverSendDir  = mean(avgGoodputOverRep, 3); % mean over sending direction
avgGoodputOverSendDir = reshape(avgGoodputOverSendDir, size(avgGoodputOverSendDir, [1 2]));

avgGoodputBetweenAllClients = reshape(avgGoodputOverSendDir, [], 1); % mean over all clients
avgGoodputBetweenAllClients(avgGoodputBetweenAllClients == 0) = [];
avgGoodput = mean(avgGoodputBetweenAllClients);
stdGoodput = std(avgGoodputBetweenAllClients, 1);

avgGoodputExport = round((avgGoodput / (1024 * 1024)), 2);
stdGoodputExport = round((stdGoodput / (1024 * 1024)), 2);

throughput  = exp.dataCubusTSharkSCION;

avgThroughputOverRep  = mean(throughput, 4); % mean over repetition
avgThroughputOverRep = reshape(avgThroughputOverRep, size(avgThroughputOverRep, [1 2 3]));

avgThroughputOverSendDir  = mean(avgThroughputOverRep, 3); % mean over sending direction
avgThroughputOverSendDir = reshape(avgThroughputOverSendDir, size(avgThroughputOverSendDir, [1 2]));

avgThroughputBetweenAllClients = reshape(avgThroughputOverSendDir, [], 1); % mean over all clients
avgThroughputBetweenAllClients(avgThroughputBetweenAllClients == 0) = [];
avgThroughput = mean(avgThroughputBetweenAllClients);
stdThroughput = std(avgThroughputBetweenAllClients, 1);

avgThroughputExport = round((avgThroughput / (1024 * 1024)), 2);
stdThroughputExport = round((stdThroughput / (1024 * 1024)), 2);

if export
    outputPath = pathToReportFolder + "/Illustrations/PerformanceEvaluation/QuicT";
        
        % Write the data for the report table
        outputMatrix = [avgGoodputExport; stdGoodputExport; ...
            avgThroughputExport; stdThroughputExport];
        writematrix(outputMatrix, outputPath);
end

end