function [hostID, remoteID, sendDir, repetition] = parseSingleIperfInfo(rawLine)

infoFields = strsplit(rawLine);

hostID          = str2num(cell2mat(infoFields(1))) + 1;
remoteID        = str2num(cell2mat(infoFields(2))) + 1;
sendDir         = str2num(cell2mat(infoFields(3))) + 1;
repetition      = str2num(cell2mat(infoFields(4)));

end

