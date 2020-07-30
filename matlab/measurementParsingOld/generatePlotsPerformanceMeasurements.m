function generatePlotsPerformanceMeasurements(exp, pathToExperiment)

    set(0,'defaulttextinterpreter','latex')

    plotBandwidthPerClientPair      = 0; % Bandwidth for each distinct client pair
    plotBandwidthPerPathSelection   = 1; % Average bandwidth per path selection
    plotSummarized                  = 1; % (Summarized for report.)

    plotMetricsDifference           = 1;

    if plotBandwidthPerClientPair 
        for side = 1:2

            yBandwidthPerPathSelectionMeanStacked = [];
            yBandwidthPerPathSelectionErrStacked  = [];

            % Per path selection, for each dinstict hostClient, remoteClient pair,
            % plot bandwidth curve for the different number of interfaces.
            for pathSelection = 1:exp.nPathSelections

                yBandwidthPerClientMeanStacked = [];

                % Check if there are any measurements
                if all(exp.dataCubus(pathSelection,:,:,:,:,:,:,:) == 0)
                  continue
                end

                for clientHostIdIndex = 1:numel(exp.clients)
                    for remoteHostIdIndex = 1:numel(exp.clients)

                       clientHostId = exp.clients(clientHostIdIndex);
                       remoteHostId = exp.clients(remoteHostIdIndex);

                       % Check if there are any measurements
                       if all(exp.dataCubus(pathSelection,clientHostId,remoteHostId,:,:,:,:,:) == 0)
                           continue
                       end

                        % Check if there are any measurements
                        if all(exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,:) == 0)
                            continue
                        end

                        % Take away the bandwidth data slice
                        metricsDataSlice = exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,2);
                        % Interfaces, Repetition, Duration
                        metricsDataSlice = reshape(metricsDataSlice, size(metricsDataSlice,5:7));

                        yBandwidthPerClientMean = reshape(mean(metricsDataSlice, 2), size(mean(metricsDataSlice, 2), [1,3]));
                        yBandwidthPerClientMeanErr  = reshape(std(metricsDataSlice, 0, 2), size(std(metricsDataSlice, 0, 2), [1,3]));
                        xTimeData = 1:exp.duration;

                        yBandwidthPerClientMeanStacked = cat(3, yBandwidthPerClientMeanStacked, yBandwidthPerClientMean);

                        % Plot bandwidth per client pair e.g. AP0 -> AP1
                        figure;

                        % Convert data into Mbits/sec
                        yBandwidthPerClientMean = yBandwidthPerClientMean / (1024 * 1024);
                        yBandwidthPerClientMeanErr = yBandwidthPerClientMeanErr / (1024 * 1024);

                        xAxis = "Time (s)";
                        yAxis = exp.dataDescription(2) + " (Mbits/sec)";

                        title = exp.clientDescription(clientHostId) + "->" + exp.clientDescription(remoteHostId) ...
                            + " (Path selection: " + exp.pathSelectionDescription(exp.pathSelections(pathSelection)) + ", Side: " ...
                            + exp.sideDescription(side) + ")";

                        plotError = 0; plotLegend = 1;
                        plotFunc1(xAxis, yAxis, xTimeData, yBandwidthPerClientMean, yBandwidthPerClientMeanErr, string(exp.interfaces), "Number of interfaces", title, plotError, plotLegend)
                   end
                end


                yBandwidthPerPathSelectionMean = mean(yBandwidthPerClientMeanStacked,3);
                yBandwidthPerPathSelectionErr  = std(yBandwidthPerClientMeanStacked, 0, 3);

                yBandwidthPerPathSelectionMeanStacked = cat(3, yBandwidthPerPathSelectionMeanStacked, yBandwidthPerPathSelectionMean);
                yBandwidthPerPathSelectionErrStacked  = cat(3, yBandwidthPerPathSelectionErrStacked, yBandwidthPerPathSelectionErr);

            end    
        end
    end

    if plotBandwidthPerPathSelection

        fig = figure;
        title = "Average bandwidth per path selection ("; 
        for index = 1:length(exp.clients)
            if index < length(exp.clients)
                title = title + exp.clientDescription(exp.clients(index)) + ", ";
            else
                title = title + exp.clientDescription(exp.clients(index));
            end
        end
        title = title + ", " + exp.nRepetition + " repetitions)";
        sgtitle(title);

        for side = 1:2

        % Per path selection, for each dinstict hostClient, remoteClient pair,
        % plot bandwidth curve for the different number of interfaces.
        for pathSelection = 1:exp.nPathSelections

            yBandwidthPerClientMeanStacked = [];

            % Check if there are any measurements
            if all(exp.dataCubus(pathSelection,:,:,:,:,:,:,:) == 0)
                continue
            end

            for clientHostIdIndex = 1:numel(exp.clients)
                for remoteHostIdIndex = 1:numel(exp.clients)

                clientHostId = exp.clients(clientHostIdIndex);
                remoteHostId = exp.clients(remoteHostIdIndex);

                % Check if there are any measurements
                if all(exp.dataCubus(pathSelection,clientHostId,remoteHostId,:,:,:,:,:) == 0)
                    continue
                end

                % Check if there are any measurements
                if all(exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,:) == 0)
                    continue
                end

                % Take away the bandwidth data slice
                metricsDataSlice = exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,2);
                % Interfaces, Repetition, Duration
                metricsDataSlice = reshape(metricsDataSlice, size(metricsDataSlice,5:7));

                yBandwidthPerClientMean = reshape(mean(metricsDataSlice, 2), size(mean(metricsDataSlice, 2), [1,3]));
                yBandwidthPerClientMeanStacked = cat(3, yBandwidthPerClientMeanStacked, yBandwidthPerClientMean);

                end
            end

            yBandwidthPerPathSelectionMean  = mean(yBandwidthPerClientMeanStacked,3);
            yBandwidthPerPathSelectionErr   = std(yBandwidthPerClientMeanStacked, 0, 3);

            yBandwidthPerPathSelectionMean  = yBandwidthPerPathSelectionMean / (1024 * 1024);
            yBandwidthPerPathSelectionErr   = yBandwidthPerPathSelectionErr / (1024 * 1024);

            xTimeData = 1:exp.duration;

            xAxis = "Time (s)";
            yAxis = exp.dataDescription(2) + " (Mbits/sec)";

            title = exp.pathSelectionDescription(exp.pathSelections(pathSelection)) + ",  measured @ " + exp.sideDescription(side);

            subplot(2, 3, (side - 1) * 3 + pathSelection)

            if (side - 1) * 3 + pathSelection == 1
                plotLegend = 1;
            else
                plotLegend = 0;
            end

            plotError = 0;
            plotFunc1(xAxis, yAxis, xTimeData, yBandwidthPerPathSelectionMean, yBandwidthPerPathSelectionErr, string(exp.interfaces), "Number of paths", title, plotError, plotLegend);

            xlim([0 exp.duration]);
            ylim([0 10.5]);

        end
        end

        fig.PaperPositionMode = 'manual';
        orient(fig,'landscape');
        print(fig,pathToExperiment+"/bandwidthPerPathSelection",'-dpdf','-loose')
    end

    if plotSummarized
        for side = 1:2       

            yBandwidthMeanPerPathSelection = [];
            yBandwidthStdPerPathSelection  = [];

            for pathSelection = 1:exp.nPathSelections

                % Check if there are any measurements
                if all(exp.dataCubus(pathSelection,:,:,:,:,:,:,:) == 0)
                  continue
                end

                yBandwidthMeanOverTimeStacked = [];

                for clientHostIdIndex = 1:numel(exp.clients)
                    for remoteHostIdIndex = 1:numel(exp.clients)

                       clientHostId = exp.clients(clientHostIdIndex);
                       remoteHostId = exp.clients(remoteHostIdIndex);

                       % Check if there are any measurements
                       if all(exp.dataCubus(pathSelection,clientHostId,remoteHostId,:,:,:,:,:) == 0)
                           continue
                       end

                        % Check if there are any measurements
                        if all(exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,:) == 0)
                            continue
                        end

                        % Take away the bandwidth data slice
                        metricsDataSlice = exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,2);
                        % Interfaces x Repetition x Duration
                        metricsDataSlice = reshape(metricsDataSlice, size(metricsDataSlice,5:7));

                        % Interfaces x Repetition
                        yBandwidthMeanOverTime = mean(metricsDataSlice, 3);
                        yBandwidthMeanOverTimeStacked = [yBandwidthMeanOverTimeStacked, yBandwidthMeanOverTime];

                    end
                end   

                yBandwidthMean = mean(yBandwidthMeanOverTimeStacked,2)';
                yBandwidthStd  = std(yBandwidthMeanOverTimeStacked,1,2)';

                yBandwidthMeanPerPathSelection = [yBandwidthMeanPerPathSelection; yBandwidthMean];
                yBandwidthStdPerPathSelection = [yBandwidthStdPerPathSelection; yBandwidthStd];

            end

            fig = figure;
            title = "Average bandwidth measured @ " + exp.sideDescription(side);

            xNumberOfPathsData = exp.interfaces;

            yBandwidthMeanPerPathSelection = yBandwidthMeanPerPathSelection / (1024 * 1024);
            yBandwidthStdPerPathSelection  = yBandwidthStdPerPathSelection / (1024 * 1024);


            xAxis = "Number of paths";
            yAxis = exp.dataDescription(2) + " (Mbits/sec)";

            plotError = 1; plotLegend = 1;
            plotFunc1(xAxis, yAxis, xNumberOfPathsData, yBandwidthMeanPerPathSelection, yBandwidthStdPerPathSelection, ...
            string(exp.pathSelectionDescription), "Path selection", title, plotError, plotLegend);

            xlim([0 max(exp.interfaces)+2]);
            xticks(0:1:max(exp.interfaces)+2)
            ylim([0 13]);

            tightInset = get(gca, 'TightInset');
            position(1) = tightInset(1);
            position(2) = tightInset(2);
            position(3) = 1 - tightInset(1) - tightInset(3);
            position(4) = 1 - tightInset(2) - tightInset(4);
            set(gca, 'Position', position);
            fig.PaperPositionMode = 'manual';
            orient(fig,'landscape');
            
            %pbaspect([1 1 1])
            %print(fig,pathToExperiment+"/summarized"+exp.sideDescription(side),'-dpdf');
            print(fig,pathToExperiment+"/summarized"+exp.sideDescription(side),'-depsc', '-loose');
            
            if exp.exportForReport && side == 2
                exportPath = exp.exportPathReport + exp.exportNameReport;
                print(fig,exportPath,'-depsc', '-loose');
            end
            
        end
    end


    % dataCubusShila(pathSelection, hostID, remoteID, nInterface, repetition, :)
    
    if plotMetricsDifference

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
            metricsDataSlice = exp.dataCubusShila(:,clientHostId,remoteHostId,:,:,:);
            % Path Selection x Interfaces x Repetition x Metrics (mtu, len, shar)
            metricsDataSlice = reshape(metricsDataSlice, size(metricsDataSlice,[1,4:6]));
  
            metricsMeanPerClientPair = mean(metricsDataSlice,3);
            
            % Path Selection x Interface x Metrics
            metricsMeanPerClientPair = reshape(metricsMeanPerClientPair, size(metricsMeanPerClientPair, [1,2,4]));
                   
            metricsDiffMTU  = (metricsMeanPerClientPair(:,:,1) - metricsMeanPerClientPair(1, :, 1)) ...
                        ./ metricsMeanPerClientPair(1, :, 1);
            metricsDiffLen  = (metricsMeanPerClientPair(:,:,2) - metricsMeanPerClientPair(2, :, 2)) ...
                        ./ metricsMeanPerClientPair(2, :, 2);
            metricsDiffShar = (metricsMeanPerClientPair(:,:,3) - metricsMeanPerClientPair(3, :, 3)) ...
                        ./ metricsMeanPerClientPair(3, :, 3);
                
            %metricsDiffMTU  = reshape(metricsDiffMTU, size(metricsDiffMTU, [1,2]));
            %metricsDiffLen  = reshape(metricsDiffLen, size(metricsDiffLen, [1,2]));
            %metricsDiffShar = reshape(metricsDiffShar, size(metricsDiffShar, [1,2]));
                    
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
    end  
    
    
    
    
    meanDiffMTU  = mean(metricsDiffsMTUStacked,3);
    meanDiffMTU  = reshape(meanDiffMTU, size(meanDiffMTU, [1,2]))
    
    stdDiffMTU   = std(metricsDiffsMTUStacked, 1, 3);
    stdDiffMTU   = reshape(stdDiffMTU, size(stdDiffMTU, [1,2]))
    
    meanDiffLen  = mean(metricsDiffsLenStacked,3);
    meanDiffLen  = reshape(meanDiffLen, size(meanDiffLen, [1,2]))
    
    stdDiffLen   = std(metricsDiffsLenStacked, 1, 3);
    stdDiffLen   = reshape(stdDiffLen, size(stdDiffLen, [1,2]))
    
    meanDiffShar = mean(metricsDiffsSharStacked,3);
    meanDiffShar  = reshape(meanDiffShar, size(meanDiffShar, [1,2]))
    
    stdDiffShar  = std(metricsDiffsSharStacked, 1, 3);
    stdDiffShar   = reshape(stdDiffShar, size(stdDiffShar, [1,2]))
    
end





