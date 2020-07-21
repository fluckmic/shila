function fig = plotFunc1(nameX, nameY, valX, valY, errY, labels, labelTitle, plotTitle, plotError, plotLegend)

nMeasurements = size(valY,1);

markerSize      = 10;
%markerEdgeColor = "red";
%markerFaceColor = "red";

for i = 1:nMeasurements
    
    if plotError 
        fig = errorbar(valX,valY(i,:),errY(i,:),'.', "MarkerSize", markerSize)
    else
        fig = plot(valX,valY(i,:), '.', "MarkerSize", markerSize)
    end
    
    hold on
end

xlabel(nameX);
ylabel(nameY);

title(plotTitle);

if plotLegend 
    lgnd = legend(labels,'Location', 'best');
    title(lgnd, labelTitle);
end
