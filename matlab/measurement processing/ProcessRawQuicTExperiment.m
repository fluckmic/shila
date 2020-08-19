%% Raw measurement processing (for quicT)
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

nDataQuicT                  = 1; % avg goodput
dataDescriptionQuicT        = ["Goodput"];
dataQuantityQuicT           = ["bits/sec"];

nDataTSharkSCION            = 1; % avg throughput
dataDescriptionTSharkSCION  = ["Throughput"];
dataQuantityTSharkSCION     = ["bits/sec"];

sideDescription             = ["Client", "Server"];

%% Parsing

% Load the experiment log to retreive necessary informations
[clients, nClients, transfer, nRepetition] = ...
    parseQuicTExperimentInfo(fullfile(pathToExperiment, "experiment.log"));

% Allocate the different cubus data cubus
% client | server | sending direction | repetition | measurement value
dataCubusQuicT          = zeros(max(clients), max(clients), 2, nRepetition, nDataQuicT);
% client | server | sending direction | repetition | measurement value
dataCubusTSharkSCION    = zeros(max(clients), max(clients), 2, nRepetition, nDataTSharkSCION);

%% Create GUI for convenience

f = uifigure;
d = uiprogressdlg(f,'Title','Processing the raw data..','Message','folder name');

% Parse the log files
RepetitionList = dir(fullfile(pathToExperiment, "/successful/", "**", "/_quicTSenderSide.log"));
for i = 1:length(RepetitionList)
    
    d.Value = i / length(RepetitionList);
    d.Message = RepetitionList(i).name;
    
    % Parse the quicT client log files
    [avgGoodput, hostID, remoteID, sendDir, repetition] = ...
        parseSingleQuicTRepetition(fullfile(RepetitionList(i).folder, RepetitionList(i).name));
    dataCubusQuicT(hostID, remoteID, sendDir, repetition, :) = avgGoodput;
    
    % Parse the scion data traffic captured with tshark
    [avgThroughput, durationCaptured] = ...
        parseSingleSCIONTrafficRepetition(fullfile(RepetitionList(i).folder, "_tsharkSCIONTraffic.csv"));
    dataCubusTSharkSCION(hostID, remoteID, sendDir, repetition, :) = avgThroughput;
    
end

close(f)

%% Create the experiment struct

exp.nDataQuicT                  = nDataQuicT;
exp.dataDescriptionQuicT        = dataDescriptionQuicT;
exp.dataQuantityQuicT           = dataQuantityQuicT;
exp.dataCubusQuicT              = dataCubusQuicT;

exp.nDataTSharkSCION            = nDataTSharkSCION;
exp.dataDescriptionTSharkSCION  = dataDescriptionTSharkSCION;
exp.dataQuantityTSharkSCION     = dataQuantityTSharkSCION;
exp.dataCubusTSharkSCION        = dataCubusTSharkSCION;

exp.nClients                    = nClients;
exp.clients                     = clients;
exp.clientDescription           = clientDescription;

exp.nRepetition                 = nRepetition;
exp.transfer                    = transfer;

%% Store the experiment struct

save(fullfile(pathToExperiment, "experiment.mat"), "-struct", "exp");