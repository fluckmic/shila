function [time, transfer, bandwidth] = parseSingleIperfLogLine(rawLine)

fields           = strsplit(rawLine);

interval         = cell2mat(fields(3));
timeEndpoints    = strsplit(interval, '-');
time             = str2num(cell2mat(timeEndpoints(2)));

transfer    = str2num(cell2mat(fields(5)));
if fields{6} ~= "Bytes"
    switch fields{6}
        case "KBytes"
            transfer = 1024 * transfer;
        case "MBytes"
            transfer = 1024 * 1024 * transfer;
        otherwise
            throw(MException("parse:NoConvertion", "No convertion for %s.", fields{6}))
    end
end

% Convert to bits/sec if not.
bandwidth   = str2num(cell2mat(fields(7)));
if fields{8} ~= "bits/sec"
    switch fields{8}
        case "Kbits/sec"
            bandwidth = 1024 * bandwidth;
        case "Mbits/sec"
            bandwidth = 1024 * 1024 * bandwidth;
        otherwise
            throw(MException("parse:NoConvertion", "No convertion for %s.", fields{8}))
    end
end
end

