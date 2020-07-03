function generatePlotsPerformanceMeasurements(exp, pathToExperiment)


% Per path selection, for each dinstict hostClient, remoteClient pair,
% plot bandwidth curve for the different number of interfaces.
for pathSelection = 1:exp.nPathSelections
   
   % Check if there are any measurements
   if all(exp.dataCubus(pathSelection,:,:,:,:,:,:,:) == 0)
       continue
   end
   
   for clientHostId = 1:exp.nClients
       for remoteHostId = 1:exp.nClients
           
           % Check if there are any measurements
           if all(exp.dataCubus(pathSelection,clientHostId,remoteHostId,:,:,:,:,:) == 0)
               continue
           end
           
           for side = 1:2
              
                % Check if there are any measurements
                if all(exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,:) == 0)
                    continue
                end
                                
                % Take away the bandwidth data slice
                dataSlice = exp.dataCubus(pathSelection,clientHostId,remoteHostId,side,:,:,:,2);
                % Interfaces, Repetition, Duration
                dataSlice = reshape(dataSlice, size(dataSlice,5:7));
                
                yData = reshape(mean(dataSlice, 2), size(mean(dataSlice, 2), [1,3]));
                yErr  = reshape(std(dataSlice, 0, 2), size(std(dataSlice, 0, 2), [1,3]));
                xData = 1:exp.duration;
                
                xAxis = "Time (s)";
                yAxis = exp.dataDescription(2) + " (" + exp.dataQuantity(2) + ")";
                
                title = exp.clientDescription(clientHostId) + "->" + exp.clientDescription(remoteHostId) ...
                    + " (Path selection: " + exp.pathSelectionDescription(exp.pathSelections(pathSelection)) + ", Side: " ...
                    + exp.sideDescription(side) + ")";
                
                plotFunc1(xAxis, yAxis, xData, yData, yErr, string(exp.interfaces), "Number of interfaces", title)
           end
       end
   end
end

end

