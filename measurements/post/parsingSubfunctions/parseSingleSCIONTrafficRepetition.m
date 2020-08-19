function [avgThroughput, duration] = parseSingleSCIONTrafficRepetition(path)

    function [second] = parseTimestamp(timestamp)
        try
            secondStrings = extractBetween(timestamp,":",".");
            secondStrings = extractAfter(string(secondStrings(1)),":");
            second        = str2num(string(secondStrings(1)));  
        catch
            second = -2;
        end
    end

M = readcell(path, "Delimiter", ",", "Range", [2 3]);

mask = cellfun(@ismissing, M(:,2));
M(mask, 2) = {0};

time = int32(cellfun(@parseTimestamp, M(:,1)));
packetSize = int32(cell2mat(M(:,2)));

perSecond = [];
currentSecond = -1; sum = 0;
for i = 1:size(time)

    if time(i) == -2
        continue
    end
    
    if time(i) ~= currentSecond
        perSecond = [perSecond; sum];
        currentSecond = time(i); sum = 0;
    end
       
    sum = sum + packetSize(i);
   
end

canditates = find(perSecond > 0.25 * max(perSecond));
perSecond = perSecond(canditates(1):canditates(end));

avgThroughput = mean(perSecond) .* 8;       % conversion to bits
duration      = size(perSecond, 1);
end

