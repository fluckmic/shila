
%% Distribution of flows
%  Generates the illustration for the report.

clear
close all

dbstop if error

pathToCapture = "~/pCloudDrive/NonCrypto Folder/02-shila/Report/Illustrations/Shila-Introduction/TestRunRaw";
[filenameCapture, pathToCapture] = uigetfile( "*.csv", pathToCapture);

data = readmatrix(fullfile(pathToCapture, filenameCapture),'NumHeaderLines',1);

% Throw away the first n lines before the actual data transfer started
data = data(5:end,:);

time = 1:length(data(:,1));

fig = figure;

set(0,'defaulttextinterpreter','latex')

linewidth = 3;
hold on
plot(time, (data(:,2) * 8) / (1024 * 1024), 'r','linewidth', linewidth);
plot(time, (data(:,3) * 8) / (1024 * 1024), "b-.",'linewidth', linewidth);
plot(time, (data(:,4) * 8) / (1024 * 1024), "m:",'linewidth', linewidth);
plot(time, (data(:,5) * 8) / (1024 * 1024), "k--",'linewidth', linewidth);

%title(descPathSelections(i), 'FontWeight', 'bold', 'FontSize', 24);
   
xlabel("Time [s]");
ylabel("Goodput [MBits/sec]");
   
lgnd = legend(["Connection", "Main-Flow", "Sub-Flow 1", "Sub-Flow 2"],'Location','best','Interpreter','latex','FontSize', 16);
   
% Cosmetics
ax = gca;
ax.XAxis.FontSize = 16;
ax.YAxis.FontSize = 16;

yticks(0:2:22);
ylim([0 22]);
xticks(0:1:max(time)+2);
xlim([0 max(time)+2]);

set(gca,'TickLabelInterpreter','latex')

tightInset = get(gca, 'TightInset');
            position(1) = tightInset(1);
            position(2) = tightInset(2);
            position(3) = 1 - tightInset(1) - tightInset(3);
            position(4) = 1 - tightInset(2) - tightInset(4);
            set(gca, 'Position', position);
            
 print(fig,"../DistributionOfFlows",'-depsc', '-loose');