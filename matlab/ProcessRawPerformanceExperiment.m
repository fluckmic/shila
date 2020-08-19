%% Raw measurement processing (for performance)
%  Processes a raw measurement and generates a struct holding all
%  the data in a condensed and nice way usable for further processing.

clear
close all

addpath outputSubfunctions/
addpath parsingSubfunctions/

dbstop if error

%% Preamble

pathToExperiment = "~/shilaExperiments";
pathToExperiment = uigetdir(pathToExperiment);

clientDescription           = ["AP0", "AP1", "AP2", "AP3"];
pathSelectionDescription    = ["MTU", "Shortest path", "Sharability"];

nDataIperf                  = 2; % transfer, bandwidth
dataDescriptionIperf        = ["Transfer", "Goodput"];
dataQuantityIperf           = ["Bytes", "bits/sec"];

nDataShila                  = 3; % avg mtu, avg len, avg shar
dataDescriptionShila        = ["Avg. MTU", "Avg. Len", "Avg. Shar"];
dataQuantityShila           = ["Bytes", "", ""];

nDataTSharkSCION            = 1; % avg throughput
dataDescriptionTSharkSCION  = ["Throughput"];
dataQuantityTSharkSCION     = ["bits/sec"];

sideDescription             = ["Client", "Server"];

%% Parsing

% Load the experiment log to retreive necessary informations
[clients, nClients, interfaces, nInterfaceCounts, pathSelections, nPathSelections, duration, nRepetition] = parseExperimentInfo(fullfile(pathToExperiment, "experiment.log"));

% Allocate the different cubus data cubus
% path selection | client | server | measurement side | sending direction | interface | repetition | time | measurement value
dataCubusIperf          = zeros(max(pathSelections), max(clients), max(clients), 2, 2, max(interfaces), nRepetition, duration, nDataIperf);
% path selection | client | server | sending direction | interface | repetition | measurement value
dataCubusShila          = zeros(max(pathSelections), max(clients), max(clients), 2, max(interfaces), nRepetition, nDataShila);
% path selection | client | server | sending direction | interface | repetition | measurement value
dataCubusTSharkSCION    = zeros(max(pathSelections), max(clients), max(clients), 2, max(interfaces), nRepetition, nDataTSharkSCION);

%% Create GUI for convenience

f = uifigure;
d = uiprogressdlg(f,'Title','Processing the raw data..','Message','folder name');

% Parse the log files
RepetitionList = dir(fullfile(pathToExperiment, "/successful/", "**", "/_iperfClientSide*"));
for i = 1:length(RepetitionList)
    
    d.Value = i / length(RepetitionList);
    d.Message = RepetitionList(i).name;
    
    % Parse the iperf client log files
    [measurementsClient, measurementsServer, pathSelection, hostID, remoteID, sendDir, nInterface, repetition] = parseSingleIperfRepetition(fullfile(RepetitionList(i).folder, RepetitionList(i).name));
    dataCubusIperf(pathSelection, hostID, remoteID, 1, sendDir, nInterface, repetition, :, :) = measurementsClient;
    dataCubusIperf(pathSelection, hostID, remoteID, 2, sendDir, nInterface, repetition, :, :) = measurementsServer;
    
    % Parse the shila client log files
    [avgMtu, avgLen, avgShar] = parseSingleShilaRepetition(fullfile(RepetitionList(i).folder, "_shilaClientSide.log"));
    dataCubusShila(pathSelection, hostID, remoteID, sendDir, nInterface, repetition, :) =  [avgMtu, avgLen, avgShar];
    
    % Parse the scion data traffic captured with tshark
    [avgThroughput, durationCaptured] = parseSingleSCIONTrafficRepetition(fullfile(RepetitionList(i).folder, "_tsharkSCIONTraffic.csv"));
    %avgThroughput
    %durationCaptured
    dataCubusTSharkSCION(pathSelection, hostID, remoteID, sendDir, nInterface, repetition, :) = avgThroughput;
end


%% Create the experiment struct

exp.nPathSelections             = nPathSelections;
exp.pathSelections              = pathSelections;
exp.pathSelectionDescription    = pathSelectionDescription;

exp.nDataIperf                  = nDataIperf;
exp.dataDescriptionIperf        = dataDescriptionIperf;
exp.dataQuantityIperf           = dataQuantityIperf;
exp.dataCubusIperf              = dataCubusIperf;

exp.nDataShila                  = nDataShila;
exp.dataDescriptionShila        = dataDescriptionShila;
exp.dataQuantityShila           = dataQuantityShila;
exp.dataCubusShila              = dataCubusShila;

exp.nDataTSharkSCION            = nDataTSharkSCION;
exp.dataDescriptionTSharkSCION  = dataDescriptionTSharkSCION;
exp.dataQuantityTSharkSCION     = dataQuantityTSharkSCION;
exp.dataCubusTSharkSCION        = dataCubusTSharkSCION;

exp.nClients                    = nClients;
exp.clients                     = clients;
exp.clientDescription           = clientDescription;

exp.nInterfaceCounts            = nInterfaceCounts;
exp.interfaces                  = interfaces;

exp.nRepetition                 = nRepetition;
exp.duration                    = duration;

exp.sideDescription             = sideDescription;

%% Store the experiment struct

save(fullfile(pathToExperiment, "experiment.mat"), "-struct", "exp");