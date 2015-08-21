# gormeter

Instructions:
jmeter -n -t <test-plan> -l <output>

e.g.
jmeter -n -t gormeter-http-10t-1000l.jmx -l ./out/gormeter-http-10t-1000l-`date +%s`.log
