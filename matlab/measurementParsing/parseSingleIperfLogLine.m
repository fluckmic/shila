function [time, transfer, bandwidth] = parseSingleIperfLogLine(rawLine)

fields           = strsplit(rawLine);

interval         = cell2mat(fields(3));
timeEndpoints    = strsplit(interval, '-');
time             = str2num(cell2mat(timeEndpoints(2)));

transfer    = str2num(cell2mat(fields(5)));
bandwidth   = str2num(cell2mat(fields(7)));

% Convert to bits/sec if not.

end

