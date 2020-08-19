%% Measurement output generation
%  Creates output from an experiment struct.

addpath outputSubfunctions/
addpath parsingSubfunctions/

clear
close all

export = 1;

%dbstop if error

%% Preample

%% Load the experiment struct

pathToExperimentStruct = "~/shilaExperiments";
[filenameExperimentStruct, pathToExperimentStruct] = uigetfile(pathToExperimentStruct);

exp = load(fullfile(pathToExperimentStruct, filenameExperimentStruct));

%% Generate output

pathToReportFolder = "~/pCloudDrive/NonCrypto Folder/02-shila/Report";

% Performance measurement data table 1 & 2 (PMDataTable1, PMDataTable2)
% +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

% Create table holding goodput and throughput for differet number of 
% paths and the different path selection metrics 
% (performance measurement data table 1 or short PMDataTable1)

% PMDataTable2 holds this data for each client server pair

[PMDataTable1, PMDataTable2] = createPMDataTable12(exp);

%plotPMDataTable1(PMDataTable1, exp.pathSelectionDescription, pathToReportFolder, export);
%plotPMDataTable2(PMDataTable2, exp.clients, exp.clientDescription, exp.pathSelectionDescription, pathToReportFolder, export);
%plotPMDataTable3(PMDataTable2, exp.clients, exp.clientDescription, exp.pathSelectionDescription, pathToReportFolder, export);

outputAvgDeviationFromOpt(exp, pathToReportFolder, export)
