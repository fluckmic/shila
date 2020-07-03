%% Parse experiment

clear
close all

clientDescription           = ["AP0", "AP1", "AP2", "AP3"];
pathSelectionDescription    = ["MTU", "Length", "Disjointness"];

nData                       = 2; % transfer, bandwidth
dataDescription             = ["Transfer", "Bandwidth"];
dataQuantity                = ["Bytes", "bits/sec"];

sideDescription             = ["Client", "Server"];

pathToExperiment = "~/go/src/shila/measurements/performance/";
pathToExperiment = uigetdir(pathToExperiment);

if ~isfile(fullfile(pathToExperiment,"experiment.mat"))
    
    % Load the experiment log to retreive necessary informations
    [clients, nClients, interfaces, nInterfaceCounts, pathSelections, nPathSelections, duration, nRepetition] = parseExperimentInfo(fullfile(pathToExperiment, "experiment.log"));

    % Generate the data cubus
    dataCubus = zeros(max(pathSelections), nClients, nClients, 2, max(interfaces), nRepetition, duration, nData); 

    % Parse the iperf Log files
    RepetitionList = dir(fullfile(pathToExperiment, "**", "_iperfClientSide*"));
    for i = 1:length(RepetitionList)
        [measurementsClient, measurementsServer, pathSelection, hostID, remoteID, nInterface, repetition] = parseSingleRepetition(fullfile(RepetitionList(i).folder, RepetitionList(i).name));

        dataCubus(pathSelection, hostID, remoteID, 1, nInterface, repetition, :, :) = measurementsClient;
        dataCubus(pathSelection, hostID, remoteID, 2, nInterface, repetition, :, :) = measurementsServer;
    end

    dataCubus = dataCubus(pathSelections,:,:,:,interfaces,:,:,:);
    
    % Create and save experiment struct

    exp.nPathSelections             = nPathSelections;
    exp.pathSelections              = pathSelections;
    exp.pathSelectionDescription    = pathSelectionDescription;

    exp.nData                       = nData;
    exp.dataDescription             = dataDescription;
    exp.dataQuantity                = dataQuantity;
    
    exp.nClients                    = nClients;
    exp.clientDescription           = clientDescription(1:nClients);

    exp.nInterfaceCounts            = nInterfaceCounts;
    exp.interfaces                  = interfaces;
    
    exp.nRepetition                 = nRepetition;
    exp.duration                    = duration;

    exp.sideDescription             = sideDescription;
    
    exp.dataCubus = dataCubus;

    save(fullfile(pathToExperiment, "experiment.mat"), "-struct", "exp");
else
    exp = load(fullfile(pathToExperiment, "experiment.mat"));
end

% Generate plots
generatePlotsPerformanceMeasurements(exp, pathToExperiment)