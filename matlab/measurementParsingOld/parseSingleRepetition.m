
function [measurementsClient, measurementsServer, PathSelect, HostID, RemoteID, nInterfaces, Repetition] = parseSingleRepetition(path)

currentLineIndex = 0;
linesToDiscard                  = [1 2 4 5 6];
linesBetweenClientAndServerData = 9;
lineWithExperimentInfo          = 3;
firstLineWithClientData         = 7;
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
        [PathSelect, HostID, RemoteID, nInterfaces, Repetition, Duration] = parseExperimentInfoIperfLogLine(currentLine);
        
        lastLineWithClientData  = firstLineWithClientData + Duration - 1;
        firstLineWithServerData = lastLineWithClientData + linesBetweenClientAndServerData + 1;
        lastLineWithServerData  = firstLineWithServerData + Duration - 1;
       
        measurementsClient = zeros(Duration,2);
        measurementsServer = zeros(Duration,2);
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

