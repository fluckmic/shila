function [avgGoodput, hostID, remoteID, sendDir, repetition] = parseSingleQuicTRepetition(path)

avgGoodput = -1;

sendDir = 1;
currentLineIndex = 0;
linesToDiscard                  = [1 2];

fid = fopen(path);
while ~feof(fid)
    
    currentLine = fgetl(fid); % read one line
    
    currentLineIndex = currentLineIndex + 1;
    
    if ismember(currentLineIndex, linesToDiscard)
        continue
    end
    
    %parse experiment info
    if currentLineIndex == 3
        [hostID, remoteID, sendDir, repetition] = parseSingleIquicTInfoLine(currentLine);
    end
    
    %parse data
    if currentLineIndex == 4
        rawLineSplit = strsplit(currentLine);
        avgGoodput = str2num(cell2mat(rawLineSplit(1)));
        avgGoodput = avgGoodput * 8 * 1024 * 1024; % convert to Bits/sec
    end
    
    %parse server data
    if currentLineIndex > 4
        return
    end
    
end
fclose(fid);

end

