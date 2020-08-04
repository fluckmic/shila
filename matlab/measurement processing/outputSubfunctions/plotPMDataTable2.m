function plotPMDataTable2(PMDataTable2, clients, clientDescription, pathSelectionDescription)

nPathSelections = 3;

set(0,'defaulttextinterpreter','latex')

        for clientHostIdIndex = 1:numel(clients)
            for serverHostIdIndex = 1:numel(clients)

                clientHostId = clients(clientHostIdIndex);
                serverHostId = clients(serverHostIdIndex);

                if clientHostId == serverHostId
                    continue
                end
                
                for pathSelection = 1:nPathSelections

                   if pathSelection > 1 
                       continue
                   end
                    
                   figure;
                   hold on

                   index = ((pathSelection-1) * 4 ) + 2;

                   nPaths       = PMDataTable2(clientHostId,serverHostId,:,1);
                   avgGoodput   = PMDataTable2(clientHostId,serverHostId,:,index)   / (1000 * 1000);
                   stdGoodput   = PMDataTable2(clientHostId,serverHostId,:,index+1) / (1000 * 1000);
                   avgThrougput = PMDataTable2(clientHostId,serverHostId,:,index+2) / (1000 * 1000);
                   stdThrougput = PMDataTable2(clientHostId,serverHostId,:,index+3) / (1000 * 1000);

                   nPaths = reshape(nPaths,size(nPaths,2,3));
                   avgGoodput = reshape(avgGoodput,size(avgGoodput,2,3));
                   stdGoodput = reshape(stdGoodput,size(stdGoodput,2,3));
                   avgThrougput = reshape(avgThrougput,size(avgThrougput,2,3));
                   stdThrougput = reshape(stdThrougput,size(stdThrougput,2,3));
                                    
                   errorbar(nPaths, avgGoodput, stdGoodput);
                   errorbar(nPaths, avgThrougput, stdThrougput);

                   titleString = clientDescription(clientHostId)+"--"+clientDescription(serverHostId)+" "+pathSelectionDescription(pathSelection);
                   title(titleString, 'FontWeight', 'bold', 'FontSize', 24);

                   xlabel("Number of paths");
                   ylabel("[MBit/s]");

                   lgnd = legend(["Goodput", "Throughput"],'Location', 'best','Interpreter','latex','FontSize', 18);

                   % Cosmetics
                   ax = gca;
                   ax.XAxis.FontSize = 18;
                   ax.YAxis.FontSize = 18;

                   xlim([0 max(nPaths)+1]);
                   xticks(0:1:max(nPaths)+1);
                end
            end
        end
end

% markerSize      = 30;
% %markerEdgeColor = "red";
% %markerFaceColor = "red";
% 
% for i = 1:nMeasurements
%     
%     if plotError 
%         fig = errorbar(valX,valY(i,:),errY(i,:),'x', "MarkerSize", markerSize, 'linewidth', 2);
%     else
%         fig = plot(valX,valY(i,:), '.', "MarkerSize", markerSize);
%     end
%     
%     hold on
% end
% 
% ax = gca;
% 
% xlabel(nameX);
% ylabel(nameY);
% 
% title(plotTitle, 'FontWeight', 'bold', 'FontSize', 24);
% 
% ax.XAxis.FontSize = 18;
% ax.YAxis.FontSize = 18;
% 
% if plotLegend 
%     lgnd = legend(labels,'Location', 'best','Interpreter','latex','FontSize', 18);
%     title(lgnd, labelTitle);
% end
% 