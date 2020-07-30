
x = linspace(0,10,15);
y = [sin(x/2); 1-sin(x/2)];
err = [0.3*ones(size(y)); 0.15*ones(size(y))];
labels = ["measurement 1", "measurement 2"];
labelTitle = "Number of interfaces";

plotTitle = "A plot";

plotFunc1("XAxis", "YAxis", x, y, err, labels, labelTitle, plotTitle);



