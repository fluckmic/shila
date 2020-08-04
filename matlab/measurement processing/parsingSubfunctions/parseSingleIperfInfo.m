function [pathSelection, hostID, remoteID, nInterface, repetition, duration, sendDir] = parseSingleIperfInfo(rawLine)

% Remove point at the very end of the line..
rawLine = rawLine(1:end-1);
infoFields = strsplit(rawLine);

pathSelection   = str2num(cell2mat(infoFields(6))) + 1;
hostID          = str2num(cell2mat(infoFields(1))) + 1;
remoteID        = str2num(cell2mat(infoFields(2))) + 1;
nInterface      = str2num(cell2mat(infoFields(9)));
repetition      = str2num(cell2mat(infoFields(5)));
duration        = str2num(cell2mat(infoFields(7)));
sendDir         = str2num(cell2mat(infoFields(10))) + 1;

end

