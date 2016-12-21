package plotgen

const generate_r = `
#!/usr/bin/env Rscript

#
# Attributed to Adrian Cockroft, https://github.com/adrianco/headroom-plot
#
chp <- function(throughput,response, cpufile, in_file, compare_file, q=0.95, qx=F, xl="Throughput",yl="Response",tl="Throughput Over Time", ml="Headroom Plot", fit=T, max=T, splits=0) {
  # remove zero throughput and response values
  nonzer <- (throughput != 0) & (response != 0)  # array of true/false
  y <- response[nonzer]
  x <- throughput[nonzer]
  # remove outliers, keep response time points inside 95% by default
  if (q != 1.0) {
      quant <- (y < quantile(y,q))
      # optionally trim throughput outliers as well
      if (qx) quant <- quant & (x < quantile(x, q))
      x <- x[quant]
      y <- y[quant]
  }
  # make histograms and record end points for scaling
  xhist <- hist(x,plot=FALSE)
  yhist <- hist(y,plot=FALSE)
  xbf <- xhist$breaks[1]                        # first
  ybf <- yhist$breaks[1]                        # first
  xbl <- xhist$breaks[length(xhist$breaks)]     # last
  ybl <- yhist$breaks[length(yhist$breaks)]     # last
  xcl <- length(xhist$counts)                   # count length
  ycl <- length(yhist$counts)                   # count length
  xrange <- c(0.0,xbl)
  yrange <- c(0.0,ybl)
  xlen <- length(x)
  # make a multi-region layout. Each number in the matrix corresponds to plot panel. O means no plot there
  #      [,1] [,2] [,3]
  # [1,]    1    3    5
  # [2,]    4    2    0
  grid_layout_matrix <- matrix(c(1,3,5,4,2,0),2,3,byrow=TRUE)
  nf <- layout(grid_layout_matrix, c(3,1), c(1,3), TRUE)
  layout.show(nf)
  # set plot margins for throughput histogram and plot it
  par(mar=c(0,4,3,0))
  barplot(xhist$counts, axes=FALSE,
      xlim=c(xcl*0.00-xbf/((xbl-xbf)/(xcl-0.5)),xcl*1.00),
      ylim=c(0, max(xhist$counts)), space=0, main=ml)
  # set plot margins for response histogram and plot it sideways
  par(mar=c(5,0,0,1))
  barplot(yhist$counts, axes=FALSE, xlim=c(0,max(yhist$counts)),
      ylim=c(ycl*0.00-ybf/((ybl-ybf)/(ycl-0.5)),ycl*1.00),
      space=0, horiz=TRUE)
  # set plot margins for throughput time series plot
  par(mar=c(2.5,1.7,3,1))
  plot(x, main=tl, cex.axis=0.8, cex.main=0.8, type="S")
  if (splits > 0) {
    step <- xlen/splits
    for(n in 0:(splits-1)) {
      lines((1+n*step):min((n+1)*step,xlen), x[(1+n*step):min((n+1)*step,xlen)], col=4+n)
    }
  }
  # set plot margins for main plot area
  par(mar=c(5,4,0,0))
  plot(x, y, xlim=xrange, ylim=yrange, xlab=xl, ylab=yl, pch=20)
  if (max) {
    # max curve
    b <- xhist$breaks
    i <- b[2] - b[1] # interval
    maxl <- list(y[b[1] < x & x <= (b[1]+i)])
    for(n in b[c(-1,-length(b))]) maxl <- c(maxl,list(y[n < x & x <= (n+i)]))
    #print(maxl)
    maxv <- unlist(lapply(maxl,max)) # apply max function to elements of list
    #print(maxv)
    #lines(xhist$mids,maxv,col=2)  # join the dots
    #staircase plot showing the range for each max response
    lines(rep(b,1,each=2)[2:(2*length(maxv)+1)],rep(maxv,1,each=2),col=3)
  }
  if (fit) {
    # fit curve, weighted to predict high throughput
    # create persistent chpfit object using <<-
    chpfit <- glm(y ~ x, inverse.gaussian, weights=as.numeric(x))
    # add fitted values to plot, sorted by throughput
    lines(x[order(x)],chpfit$fitted.values[order(x)],col="steelblue1")
  }
  if (splits > 0) {
    step <- xlen/splits
    for(n in 0:(splits-1)) {
      Sys.sleep(1)
      points(x[(1+n*step):min((n+1)*step,xlen)],y[(1+n*step):min((n+1)*step,xlen)], xlim=xrange, ylim=yrange, col=4+n)
    }
  }
  legend <- c(basename(in_file))
  if (!is.na(compare_file)) {
    compare_data = read.delim(compare_file, header=T, sep=",")
    plot_old_fit(compare_data$throughput, compare_data$latency)
    legend <- c(legend,basename(compare_file))
  }
  legend("bottomright", inset=0.05, legend, lty=c(1,1), col=c("steelblue1", "red"))
  # plot the cpu portion in panel 5
  if (!is.na(cpufile)) {
    plot_cpu(cpufile)
  }
}

plot_cpu <- function(file){
  data <- read.csv(file)

  percentage1 <- data[[2]]
  percentage1Vector <- c(0,percentage1)

  allData <- length(data)
  y_range <-c()
  for(i in 2:allData){
    y_range <- c(y_range,data[[i]])
  }
  yrange <- range(0,y_range)
  par(mar=c(2.5,1.7,3,1))
  plot(percentage1Vector, main="CPU over Time", cex.axis=0.8, cex.main=0.8, type="n",xlab="timestamp", ylab="percentage")

  legendNames <-c()
  legendColours <-c()
  # exclude the 1st column which is timestamp
  for(i in 2:allData){
    legendNames<- c(legendNames, paste("cpu",i-1,sep=""))
    legendColours <- c(legendColours, i)
    percentageVector <- c(0,data[[i]])
    # change type to o if you need points
    lines(percentageVector, type="l", col=i)
  }

  # plot title
  # no need to display legend if single cpu sample provided
  if (allData > 2){
    xcor <- length(data$timestamp) - 5
    ycor <- max(yrange)
    legend(xcor, ycor, legendNames, cex=0.8, col=legendColours, lty=1, bty="n")
  }
}

plot_old_fit <- function(throughput, response, q=0.95, qx=F) {
  nonzer <- (throughput != 0) & (response != 0)  # array of true/false
  y <- response[nonzer]
  x <- throughput[nonzer]
  # remove outliers, keep response time points inside 95% by default
  if (q != 1.0) {
    quant <- (y < quantile(y,q))
    # optionally trim throughput outliers as well
    if (qx) quant <- quant & (x < quantile(x, q))
    x <- x[quant]
    y <- y[quant]
  }
  # fit curve, weighted to predict high throughput
  # create persistent chpfit object using <<-
  chpfit <- glm(y ~ x, inverse.gaussian, weights=as.numeric(x))
  # add fitted values to plot, sorted by throughput
  lines(x[order(x)],chpfit$fitted.values[order(x)],col="red")
}

generate <- function(in_file, out_file, compare_file, cpu_file) {
  input_data = read.delim(in_file, header=T, sep=",")
  compare_data = NA

  png(out_file, width=1200, height=1200, res=120)
  chp(input_data$throughput, input_data$latency, cpu_file, in_file, compare_file)
  dev.off()
}

handle_args <- function(){
  args = commandArgs(trailingOnly=TRUE)
  total_args <- length(args)
  usage <- "Usage: Rscript generate.r input_data.csv output.png [-comparefile compare_data.csv] [-cpufile cpu_data.csv]"
  success = FALSE
  if (total_args < 2) {
    stop(usage)
  }
  if (total_args == 2){
    success = TRUE
    generate(args[1], args[2], NA, NA)
  }
  if (total_args == 4){
    if (args[3] == "-comparefile"){
      if(!is.na(args[4])) {
        success = TRUE
        generate(args[1], args[2], args[4], NA)
      }
    }
    if (args[3] == "-cpufile"){
      if(!is.na(args[4])) {
        success = TRUE
        generate(args[1], args[2], NA, args[4])
      }
    }
  }
  if (total_args == 6){
    if (args[3] == "-comparefile" && args[5] == "-cpufile"){
      if(!is.na(args[4]) && !is.na(args[6])) {
        success = TRUE
        generate(args[1], args[2], args[4], args[6])
      }
    }
    if (args[3] == "-cpufile" && args[5] == "-comparefile"){
      if(!is.na(args[4]) && !is.na(args[6])) {
        success = TRUE
        generate(args[1], args[2], args[6], args[4])
      }
    }
  }
  if (!success){
    stop(usage)
  }
}

handle_args()
`
