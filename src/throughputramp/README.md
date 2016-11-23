# Throughputramp

##Build throughput ramp from source:

```
cd throughputramp/
go build .
./throughputramp -access-key-id [aws-blob-access-key] -secret-access-key [aws-blob-secret-access-key] -bucket-name routing-perf-graphs -n 10000 -q 100 -lower-concurrency 1 -upper-concurrency 60 -s3-region us-east-1   http://10.0.1.5:80
```

Note:
Using `-s3-endpoint` currenlty results in AWS API error, you can use `-s3-region us-east-1` instead.
