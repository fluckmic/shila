function [PathSelect, HostID, RemoteID, nInterfaces, Repetition, Duration] = parseExperimentInfoIperfLogLine(rawLine)

% Remove point at the very end of the line..
rawLine = rawLine(1:end-1);
infoFields = strsplit(rawLine);

PathSelect  = str2num(cell2mat(infoFields(6))) + 1;
HostID      = str2num(cell2mat(infoFields(1))) + 1;
RemoteID    = str2num(cell2mat(infoFields(2))) + 1;
nInterfaces = str2num(cell2mat(infoFields(9)));
Repetition  = str2num(cell2mat(infoFields(5)));
Duration    = str2num(cell2mat(infoFields(7)));

end
