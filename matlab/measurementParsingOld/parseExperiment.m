%% Parse experiment

clear
close all

exportForReport             = 1;
exportPathReport            = "~/pCloudDrive/NonCrypto Folder/02-shila/Report/Illustrations/PerformanceEvaluation/";
%exportNameReport            = "BandwidthMeasurementSummarizedCUB.eps";
exportNameReport            = "BandwidthMeasurementSummarizedLIA.eps";

clientDescription           = ["AP0", "AP1", "AP2", "AP3"];
pathSelectionDescription    = ["MTU", "Shortest path", "Sharability"];

nData                       = 2; % transfer, bandwidth
dataDescription             = ["Transfer", "Bandwidth"];
dataQuantity                = ["Bytes", "bits/sec"];

nDataShila                  = 3; % avg mtu, avg len, avg shar
dataDescriptionShila        = ["Avg. MTU", "Avg. Len", "Avg. Shar"];
dataQuantityShila           = ["Bytes", "", ""];


sideDescription             = ["client", "server"];

pathToExperiment = "~/pCloudDrive/NonCrypto Folder/02-shila/Experiments/Performance/congestion-cubic/";
pathToExperiment = uigetdir(pathToExperiment);

if ~isfile(fullfile(pathToExperiment,"experiment.mat"))
    
    % Load the experiment log to retreive necessary informations
    [clients, nClients, interfaces, nInterfaceCounts, pathSelections, nPathSelections, duration, nRepetition] = parseExperimentInfo(fullfile(pathToExperiment, "experiment.log"));

    % Generate the data cubus
    dataCubus      = zeros(max(pathSelections), max(clients), max(clients), 2, max(interfaces), nRepetition, duration, nData); 
    dataCubusShila = zeros(max(pathSelections), max(clients), max(clients), max(interfaces), nRepetition, nDataShila);  
    
    % Parse the iperf Log files
    RepetitionList = dir(fullfile(pathToExperiment, "**", "_iperfClientSide*"));
    for i = 1:length(RepetitionList)
        
        [measurementsClient, measurementsServer, pathSelection, hostID, remoteID, nInterface, repetition] = parseSingleRepetition(fullfile(RepetitionList(i).folder, RepetitionList(i).name));
        dataCubus(pathSelection, hostID, remoteID, 1, nInterface, repetition, :, :) = measurementsClient;
        dataCubus(pathSelection, hostID, remoteID, 2, nInterface, repetition, :, :) = measurementsServer;

        [avgMtu, avgLen, avgShar] = parseSingleRepetitionShilaLog(fullfile(RepetitionList(i).folder, "_shilaClientSide.log"));
        dataCubusShila(pathSelection, hostID, remoteID, nInterface, repetition, :) =  [avgMtu, avgLen, avgShar];
       
    end

    dataCubus = dataCubus(pathSelections,:,:,:,interfaces,:,:,:);
    
    % Create and save experiment struct

    exp.nPathSelections             = nPathSelections;
    exp.pathSelections              = pathSelections;
    exp.pathSelectionDescription    = pathSelectionDescription;

    exp.nData                       = nData;
    exp.dataDescription             = dataDescription;
    exp.dataQuantity                = dataQuantity;
    
    exp.nDataShila                  = nDataShila;
    exp.dataDescriptionShila        = dataDescriptionShila;
    exp.dataQuantityShila           = dataQuantityShila;
    
    exp.nClients                    = nClients;
    exp.clients                     = clients;
    exp.clientDescription           = clientDescription;

    exp.nInterfaceCounts            = nInterfaceCounts;
    exp.interfaces                  = interfaces;
    
    exp.nRepetition                 = nRepetition;
    exp.duration                    = duration;

    exp.sideDescription             = sideDescription;
    
    exp.exportForReport             = exportForReport;
    exp.exportPathReport            = exportPathReport;
    exp.exportNameReport            = exportNameReport;
    
    
    exp.dataCubus       = dataCubus;
    exp.dataCubusShila  = dataCubusShila;

    save(fullfile(pathToExperiment, "experiment.mat"), "-struct", "exp");
else
    exp = load(fullfile(pathToExperiment, "experiment.mat"));
end

% Generate plots
generatePlotsPerformanceMeasurements(exp, pathToExperiment)