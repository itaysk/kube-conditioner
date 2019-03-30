# Prometheus Datasource

This datasource provides integration with Prometheus by executing a query which is expected to return an [instant vector](https://prometheus.io/docs/prometheus/latest/querying/basics/#instant-vector-selectors). **The condition evaluation result will be true if the resulting vector is not empty.** This logic complements the general intuition of working with alerts:  
Alerts are fired when "bad things" happen, so the query of an alert is negative by nature - it is assumed to be falsy under normal operations, and truthy when something breaks. Conditions are supposed to signal some kind of health status and therefore the query that represents a condition should be positive by nature - it is assumed to be falsy in normal operations, and truthy when something breaks.

Let's review a couple of examples:

`up{job="myapp"} == 1`  
The `up` query will return an instant vector that contains all the targets known to Prometheus, with values of 1/0 to indicate their health. By adding `{job="myapp"}` we scope the result to contain only the targets of our application. By adding ` == 1` we are scoping it further by containing only the healthy targets of our application. In other words, the result of `up{job="myapp"} == 1` is expected to contain a value under normal operation, and to be empty when the target is unhealthy. Consequently, the condition will be true if the target is healthy (there's a result) and false if it isn't  (the result is empty).

`rate(httpErrors)[5m] < 5`  
Let's assume that `httpErrors` query will return an instant vector with a single counter. By adding `rate()[5m]` we turn the counter into a rate of errors over the past 5 minute window. By adding ` < 5` we are scoping it to values where the rate is lower then five. In this example, the result will contain a single value if the rate is lower then 5 (which is probably a good thing) and if the rate is greater then five the query result will be empty. Consequently, the condition will be true if there are "little" errors and false if there are "many" errors.