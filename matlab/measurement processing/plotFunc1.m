function fig = plotFunc1(nameX, nameY, valX, valY, errY, labels, labelTitle, plotTitle, plotError, plotLegend)

nMeasurements = size(valY,1);

markerSize      = 30;
%markerEdgeColor = "red";
%markerFaceColor = "red";

for i = 1:nMeasurements
    
    if plotError 
        fig = errorbar(valX,valY(i,:),errY(i,:),'x', "MarkerSize", markerSize, 'linewidth', 2);
    else
        fig = plot(valX,valY(i,:), '.', "MarkerSize", markerSize);
    end
    
    hold on
end

ax = gca;

xlabel(nameX);
ylabel(nameY);

title(plotTitle, 'FontWeight', 'bold', 'FontSize', 24);

ax.XAxis.FontSize = 18;
ax.YAxis.FontSize = 18;

if plotLegend 
    lgnd = legend(labels,'Location', 'best','Interpreter','latex','FontSize', 18);
    title(lgnd, labelTitle);
end
