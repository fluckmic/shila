function generatePlotsPerformanceMeasurements(exp, pathToExperiment)

set(0,'defaulttextinterpreter','latex')

plotBandwidthPerClientPair      = 0; % Bandwidth for each distinct client pair
plotBandwidthPerPathSelection   = 1; % Average bandwidth per path selection

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
                    bandwidthDataSlice = exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,2);
                    % Interfaces, Repetition, Duration
                    bandwidthDataSlice = reshape(bandwidthDataSlice, size(bandwidthDataSlice,5:7));

                    yBandwidthPerClientMean = reshape(mean(bandwidthDataSlice, 2), size(mean(bandwidthDataSlice, 2), [1,3]));
                    yBandwidthPerClientMeanErr  = reshape(std(bandwidthDataSlice, 0, 2), size(std(bandwidthDataSlice, 0, 2), [1,3]));
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
    title = "Average bandwidth per path selection ( "; 
    for index = 1:length(exp.clients)
        if index < length(exp.clients)
            title = title + exp.clientDescription(exp.clients(index)) + ", ";
        else
            title = title + exp.clientDescription(exp.clients(index));
        end
    end
    title = title + ", " + exp.nRepetition + " repetitions)";
    sgtitle(title)
    
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
            bandwidthDataSlice = exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,2);
            % Interfaces, Repetition, Duration
            bandwidthDataSlice = reshape(bandwidthDataSlice, size(bandwidthDataSlice,5:7));

            yBandwidthPerClientMean = reshape(mean(bandwidthDataSlice, 2), size(mean(bandwidthDataSlice, 2), [1,3]));
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
        plotFunc1(xAxis, yAxis, xTimeData, yBandwidthPerPathSelectionMean, yBandwidthPerPathSelectionErr, string(exp.interfaces), "Number of interfaces", title, plotError, plotLegend);
      
        xlim([0 exp.duration])
        ylim([0 10.5])
        
    end
    end
    
    fig.PaperPositionMode = 'manual';
    orient(fig,'landscape')
    print(fig,pathToExperiment+"/bandwidthPerPathSelection",'-dpdf','-fillpage')
end

end

