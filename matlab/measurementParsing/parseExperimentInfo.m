function [clients, nClients, interfaces, nInterfaces, pathSelections, nPathSelections, duration, nRepetition] = parseExperimentInfo(path)

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
    
    % parse interfaces
    if currentLineIndex == 5
        interfaces  = strsplit(currentLine);
        interfaces  = cellfun(@str2num, interfaces(2:end-1));
        nInterfaces = length(interfaces); 
    end
    
    % parse path selection
    if currentLineIndex == 6
        pathSelections  = strsplit(currentLine);
        pathSelections  = cellfun(@str2num, pathSelections(3:end-1)) + 1;
        nPathSelections = length(pathSelections); 
    end
    
    % parse number of repetitions
    if currentLineIndex == 7
        nRepetition  = strsplit(currentLine);
        nRepetition  = cellfun(@str2num, nRepetition(2));
    end

    % parse duration
    if currentLineIndex == 8
        duration  = strsplit(currentLine);
        duration  = cellfun(@str2num, duration(2));
    end
    
    if currentLineIndex > 8
        return
    end
    
end
fclose(fid);


end

