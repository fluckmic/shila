function plotFunc1(nameX, nameY, valX, valY, errY, labels, labelTitle, plotTitle)

figure;

nMeasurements = size(valY,1);

markerSize      = 10;
%markerEdgeColor = "red";
%markerFaceColor = "red";

for i = 1:nMeasurements
    %errorbar(valX,valY(i,:),errY(i,:),'.', "MarkerSize", markerSize)
    plot(valX,valY(i,:), '.', "MarkerSize", markerSize)
    hold on
end

xlabel(nameX);
ylabel(nameY);

title(plotTitle);

lgnd = legend(labels,'Location', 'best');
title(lgnd, labelTitle);