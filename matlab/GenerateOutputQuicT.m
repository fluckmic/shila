%% Measurement output generation (quicT)
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

% Output for quicT measurements
% +++++++++++++++++++++++++++++

outputAvgGoodAndThroughputQuicT(exp, pathToReportFolder, export)