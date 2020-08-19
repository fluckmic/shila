function outputAvgDeviationFromOpt(exp, pathToReportFolder, export)

metricsDiffsMTUStacked  = [];
metricsDiffsLenStacked  = [];
metricsDiffsSharStacked = [];

for clientHostIdIndex = 1:numel(exp.clients)
    for remoteHostIdIndex = 1:numel(exp.clients)
        
        clientHostId = exp.clients(clientHostIdIndex);
        remoteHostId = exp.clients(remoteHostIdIndex);
        
        % Check if there are any measurements
        if all(exp.dataCubusShila(:,clientHostId,remoteHostId,:,:,:) == 0)
            continue
        end
        
        
        % Take away the metrics data slice
        % path selection | client | server | sending direction | interface | repetition | measurement value
        metricsDataSlice = exp.dataCubusShila(:,clientHostId,remoteHostId,:,:,:,:);
        metricsDataSlice = reshape(metricsDataSlice, size(metricsDataSlice,[1,4:7]));
        
        metricsMeanPerClientPair = mean(metricsDataSlice,4); % mean over repetition
        metricsMeanPerClientPair = mean(metricsMeanPerClientPair,2); % sending direction
        
        % Path Selection x Interface x Metrics
        metricsMeanPerClientPair = reshape(metricsMeanPerClientPair, size(metricsMeanPerClientPair, [1,3,5]));
        
        optMTU = metricsMeanPerClientPair(1, :, 1);
        optLen = metricsMeanPerClientPair(2, :, 2);
        optSha = metricsMeanPerClientPair(3, :, 3);
        
        allMTU = metricsMeanPerClientPair(:,:,1);
        allLen = metricsMeanPerClientPair(:,:,2);
        allSha = metricsMeanPerClientPair(:,:,3);
        
        % All - Opt / Opt
        metricsDiffMTU  = (allMTU - optMTU) ./ optMTU;
        metricsDiffLen  = (allLen - optLen) ./ optLen;
        metricsDiffShar = (allSha - optSha) ./ optSha;
        
        metricsDiffMTU(isnan(metricsDiffMTU)) = 0;
        metricsDiffLen(isnan(metricsDiffLen)) = 0;
        metricsDiffShar(isnan(metricsDiffShar)) = 0;
        
        metricsDiffMTU  = metricsDiffMTU(:, exp.interfaces);
        metricsDiffLen = metricsDiffLen(:, exp.interfaces);
        metricsDiffShar = metricsDiffShar(:, exp.interfaces);
        
        metricsDiffsMTUStacked = cat(3, metricsDiffsMTUStacked, metricsDiffMTU');
        metricsDiffsLenStacked = cat(3, metricsDiffsLenStacked, metricsDiffLen');
        metricsDiffsSharStacked = cat(3, metricsDiffsSharStacked, metricsDiffShar');
        
    end
end

meanDiffMTU  = round(mean(metricsDiffsMTUStacked,3),2);
stdDiffMTU   = round(std(metricsDiffsMTUStacked, 1, 3),2);

meanDiffLen  = round(mean(metricsDiffsLenStacked,3),2);
stdDiffLen   = round(std(metricsDiffsLenStacked, 1, 3),2);

meanDiffShar = round(mean(metricsDiffsSharStacked,3),2);
stdDiffShar  = round(std(metricsDiffsSharStacked, 1, 3),2);

if export
    outputPath = pathToReportFolder + "/Illustrations/PerformanceEvaluation/AvgDev";
        
        % Write the data for the report table
        outputMatrix = [meanDiffMTU(:,1), meanDiffLen(:,1), meanDiffShar(:,1); ...
                        stdDiffMTU(:,1), stdDiffLen(:,1), stdDiffShar(:,1)] ;
        writematrix(outputMatrix, outputPath + "MTU");
        
        outputMatrix = [meanDiffMTU(:,2), meanDiffLen(:,2), meanDiffShar(:,2); ...
                        stdDiffMTU(:,2), stdDiffLen(:,2), stdDiffShar(:,2)] ;
        writematrix(outputMatrix, outputPath + "PathLength");
        
        outputMatrix = [meanDiffMTU(:,3), meanDiffLen(:,3), meanDiffShar(:,3); ...
                        stdDiffMTU(:,3), stdDiffLen(:,3), stdDiffShar(:,3)] ;
        writematrix(outputMatrix, outputPath + "Sharability");
end


end