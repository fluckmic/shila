function plotPMDataTable1(PMDataTable1, descPathSelections)

nPathSelections = 3;

set(0,'defaulttextinterpreter','latex')

for i = 1:nPathSelections
   
   figure;
   hold on
   
   index = ((i-1) * 4 ) + 2;
   
   nPaths       = PMDataTable1(:,1);
   avgGoodput   = PMDataTable1(:,index)   / (1000 * 1000);
   stdGoodput   = PMDataTable1(:,index+1) / (1000 * 1000);
   avgThrougput = PMDataTable1(:,index+2) / (1000 * 1000);
   stdThrougput = PMDataTable1(:,index+3) / (1000 * 1000);
   
   errorbar(nPaths, avgGoodput, stdGoodput);
   errorbar(nPaths, avgThrougput, stdThrougput);
   
   % 
   title(descPathSelections(i), 'FontWeight', 'bold', 'FontSize', 24);
   
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

end

