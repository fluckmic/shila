function [clients, nClients, transfer, nRepetition] = parseQuicTExperimentInfo(path)

fid = fopen(path);

currentLineIndex = 0;
while ~feof(fid) 

    currentLine = fgetl(fid); % read one line
    currentLineIndex = currentLineIndex + 1;

    % parse clients
    if currentLineIndex == 4
        clients     = strsplit(currentLine);
        clients     = cellfun(@str2num, clients(3:end-1)) + 1;
        nClients    = length(clients);
    end
        
    % parse number of repetitions
    if currentLineIndex == 5
        nRepetition  = strsplit(currentLine);
        nRepetition  = cellfun(@str2num, nRepetition(2));
    end

    % parse duration
    if currentLineIndex == 6
        transfer  = strsplit(currentLine);
        transfer  = cellfun(@str2num, transfer(2));
    end
    
    if currentLineIndex > 6
        return
    end
    
end
fclose(fid);


end

