
function [measurementsClient, measurementsServer, pathSelection, hostID, remoteID, sendDir, nInterface, repetition] = parseSingleIperfRepetition(path)

sendDir = 1;
currentLineIndex = 0;
linesToDiscard                  = [1 3 4 5 6];
linesBetweenClientAndServerData = 9;
lineWithExperimentInfo          = 2;
firstLineWithClientData         = 6;
lastLineWithClientData          = 0;
firstLineWithServerData         = 0;
lastLineWithServerData          = 0;

fid = fopen(path);
while ~feof(fid)
    
    currentLine = fgetl(fid); % read one line
    
    currentLineIndex = currentLineIndex + 1;
    
    if ismember(currentLineIndex, linesToDiscard)
        continue
    end
    
    %parse experiment info
    if currentLineIndex == lineWithExperimentInfo
        [pathSelection, hostID, remoteID, nInterface, repetition, duration, sendDir] = parseSingleIperfInfo(currentLine);
        
        firstLineWithClientData = firstLineWithClientData + (sendDir - 1);
        lastLineWithClientData  = firstLineWithClientData + duration - 1;
        firstLineWithServerData = lastLineWithClientData + linesBetweenClientAndServerData + 1;
        lastLineWithServerData  = firstLineWithServerData + duration - 1;
        
        measurementsClient = zeros(duration,2);
        measurementsServer = zeros(duration,2);
    end
    
    %parse client data
    if currentLineIndex >= firstLineWithClientData && currentLineIndex <= lastLineWithClientData
        [~, transfer, bandwidth] = parseSingleIperfLogLine(currentLine);
        time = currentLineIndex - firstLineWithClientData + 1;
        measurementsClient(time, :, :) = [transfer, bandwidth];
    end
    
    %parse server data
    if currentLineIndex >= firstLineWithServerData && currentLineIndex <= lastLineWithServerData
        time = currentLineIndex - firstLineWithServerData + 1;
        [~, transfer, bandwidth] = parseSingleIperfLogLine(currentLine);
        measurementsServer(time, :, :) = [transfer, bandwidth];
    end
    
end
fclose(fid);

end

