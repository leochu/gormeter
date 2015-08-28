# gormeter

Instructions:
jmeter -n -t `test-plan-path` -l `output-path`

### JMX Parameters

- `num_loops`: the number of loops that JMeter should run.
- `num_threads`: the number of threads that JMeter should run.
- `protocol`: the protocol of the HTTPSampler (http or https)
- `keep_alive`: keeps the connection alive through the loops
- `host`: The host to specify in the `Host:` header
- `domain`: The domain that JMeter will attempt to connect to.
- `port`: The port that JMeter will attempt to connect to.

e.g.
```jmeter -n -t gormeter-http-10t-1000l.jmx -Jnum_loops=100 -Jnum_threads=100 -l ./out/gormeter-http-10t-1000l.log```
