function plotPMDataTable1(PMDataTable1, descPathSelections, ...
    outputFolder, export)

nPathSelections = 3;
outputUnit = "MBits/sec";

set(0,'defaulttextinterpreter','latex')

for i = 1:nPathSelections
    
    index = ((i-1) * 4 ) + 2;
    
    nPaths       = PMDataTable1(:,1);
    
    avgGoodput   = PMDataTable1(:,index);
    stdGoodput   = PMDataTable1(:,index+1);
    
    avgThroughput = PMDataTable1(:,index+2);
    stdThroughput = PMDataTable1(:,index+3);
    
    avgOverhead = avgThroughput - avgGoodput;
    stdOverhead = sqrt(stdThroughput .^ 2 + stdGoodput .^ 2);
    
    if outputUnit == "MBits/sec"
        avgGoodput    = avgGoodput ./ (1024 * 1024);
        stdGoodput    = stdGoodput ./ (1024 * 1024);
        avgThroughput = avgThroughput ./ (1024 * 1024);
        stdThroughput = stdThroughput ./ (1024 * 1024);
        avgOverhead   = avgOverhead ./ (1024 * 1024);
        stdOverhead   = stdOverhead ./ (1024 * 1024);
    end
            
    fig = figure;
    hold on
    set(0,'defaulttextinterpreter','latex')
    
    linewidth = 3;
    
    errorbar(nPaths, avgGoodput , stdGoodput, ...
        'b.','linewidth', linewidth / 2, 'CapSize', 18);
    p = plot(nPaths, avgGoodput, 'b.--','linewidth', linewidth);
    p.Marker = "o";
    p.MarkerSize = 10;
    %errorbar(nPaths, avgThroughput ./ (1024 * 1024), stdThroughput ./ (1024 * 1024));
    
    %title(descPathSelections(i), 'FontWeight', 'bold', 'FontSize', 24);
     
    grid on
    
    xlabel("Number of paths");
    ylabel("Goodput [MBits/sec]");
    
    %lgnd = legend(["Goodput", "Throughput"],'Location', 'best','Interpreter','latex','FontSize', 18);
    
    % Cosmetics
    ax = gca;
    ax.XAxis.FontSize = 16;
    ax.YAxis.FontSize = 16;
    
    xlim([0 max(nPaths)+1]);
    xticks(0:1:max(nPaths)+1);
    
    limY = ylim;
    ylim([0 limY(2)]);

    set(gca,'TickLabelInterpreter','latex')
    
    tightInset = get(gca, 'TightInset');
    position(1) = tightInset(1);
    position(2) = tightInset(2);
    position(3) = 1 - tightInset(1) - tightInset(3);
    position(4) = 1 - tightInset(2) - tightInset(4);
    set(gca, 'Position', position);
    
    if export
        
        outputPath = outputFolder+"/Illustrations/PerformanceEvaluation/PerformanceWrtPath"+strrep(descPathSelections(i),' ','');
        print(fig, outputPath, '-depsc', '-loose');
        
        % Write the data for the report table
        outputMatrix = round([nPaths; avgGoodput; stdGoodput; ...
            avgThroughput; stdThroughput; avgOverhead; stdOverhead],2);
        writematrix(outputMatrix, outputPath);
        
    end
end