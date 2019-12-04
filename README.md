# HTTP log monitoring console program

The application consumes an actively written-to w3c-formatted HTTP access log (https://www.w3.org/Daemon/User/Config/Logging.html) from a specified file (default to /tmp/access.log).

It displays stats about the traffic every 10 seconds and monitors if the traffic from last 2 minutes exceeds the specified threshold. If the threshold is reached the application displays a message saying that “High traffic generated an alert - hits = {value}, triggered at {time}”. Whenever the total traffic drops again below that value on average for the past 2 minutes, the application displays a new message. 

Known limitations:
- TODO

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
