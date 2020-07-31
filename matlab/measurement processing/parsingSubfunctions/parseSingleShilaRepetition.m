function [avgMtu, avgLen, avgShar] = parseSingleShilaRepetition(path)

currentLineIndex = 0;

mtus  = [];
lens  = [];
shar = [];


fid = fopen(path);
while ~feof(fid) 
    
    currentLine = fgetl(fid); % read one line
    currentLineIndex = currentLineIndex + 1;
    
    currentLineSplitted = strsplit(currentLine);
    
    if size(currentLineSplitted,2) == 4 && currentLineSplitted(3) == "Sharability:"
        shar = [shar str2num(cell2mat(currentLineSplitted(4)))];
        continue
    end
    
    if size(currentLineSplitted,2) == 7 && currentLineSplitted(3) == "Metrics:"
        mtus = [mtus str2num(cell2mat(currentLineSplitted(4)))];
        lens = [lens str2num(cell2mat(currentLineSplitted(6)))];
        continue    
    end    
end

avgMtu  = mean(mtus);
avgLen  = mean(lens);
avgShar = mean(shar);

end

