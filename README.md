# HTTP log monitoring console program

The application consumes an actively written-to w3c-formatted HTTP access log (https://www.w3.org/Daemon/User/Config/Logging.html) from a specified file (default to /tmp/access.log) and stores them into a timeseries database.

It displays stats about the traffic every 10 seconds and monitors if the traffic from last 2 minutes exceeds the specified threshold. If the threshold is reached the application displays a message saying that “High traffic generated an alert - hits = {value}, triggered at {time}”. Whenever the total traffic drops again below that value on average for the past 2 minutes, the application displays a new message. 

Known issues:
- Testing coverage is not 100%. We didn't cover the error statements through the tests. To do that, we could create an interface for the database, generate some mocks and use those to simultate the failure. Also, the "Run" methods were not covered.
- Some errors are not treated accordingly. We should add more context and we should also cover the errors returned by goroutines though error channels. 
- The alerts should send some signals when the state is changed, in case if we would like to trigger other tasks (e.g e-mail notifications, etc). A different approach would be to configure the notifications directly on the alerts.
-

## How to run the application
In order to run the application locally:
```
go run main.go --filename=/tmp/test.log --threshold=5
```

You can also test the application using Docker. The below command starts in background a logging generator and the monitoring application. 
```
docker-compose up
```

## How to run the tests
In order to run the tests use the following command:
```
docker-compose -f docker-compose-testing.yml up
```
