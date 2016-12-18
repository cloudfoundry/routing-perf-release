#!/usr/bin/env Rscript

source("chp.r")

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
  plot(percentage1Vector, type="n",ylim=yrange, xlab="timestamp", ylab="percentage")

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
  #title(main="CPU percentage(s)", col.main="blue")
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
  chp(input_data$throughput, input_data$latency)
  legend <- c(basename(in_file))
  if (!is.na(compare_file)) {
    compare_data = read.delim(compare_file, header=T, sep=",")
    plot_old_fit(compare_data$throughput, compare_data$latency)
    legend <- c(legend,basename(compare_file))
  }
  legend("bottomright", inset=0.02, legend, lty=c(1,1), col=c("steelblue1", "red"))
  if (!is.na(cpu_file)) {
    plot_cpu(cpu_file)
  }
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
