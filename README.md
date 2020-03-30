Uses Go 1.11 modules!!

### Description ###

The app prints an alert message if the number of requests observed in the log in the last 2 minutes exceeds 10. It also prints minimal stats every 10 seconds:
- list of sections with number of hits in the last 10 seconds
- "hot section" meaning the most active section in the last 10 seconds

### In Depth ###

`alerts/` contains the Alert struct which can be configured with a #of requests / time interval that would trigger the alert.
`reporter/` is used to gather stats. Currently the only stats gathered are the number of hits per section per time interval. It can be extended to use more info from the access log 
`monitor/` exposes a Watch() method that starts checking for changes to the log file at e configurable cadence. 
The approach is to check for changes in the file size and remember last position it read from. It does all this in a separate go-routine and it has a "subscription" mechanism to send updates.
`parser/` exposes interfaces for a log parser and a log. At the moment we only have access log parser implementation but this can be extended to other types of logs and used with the file monitor/scanner

### Make targets ###
- `make run` should start the app with the default `/var/log/access.log` as the input file
- `make run-test` will start the tool with `./testing/access.log` as the file to tail
- `make run-integration` starts an "integration test" that prints logs in a random order to `./testing/access.log`

### Todo ###

1. Overall improvements:
- The app is not very friendly to the user; it is lacking a readable UI. I have attempted to use https://github.com/jroimartin/gocui but don't have a working version yet
- inotify could be used to observe changes to the log file

1. Alerts package 
- Alert is not quite thread safe. It needs some work to get there and as a result increasing test coverage would be easier too
- Alert could use a function or interface that is called when an alert is triggered or "resolved"
- error handling needs attention
- could use a logger passed in

2. Reporter package
- similar to alerts, this is not quite thread safe
- test coverage needs some increasing as well
- error handling needs attention
- could use a logger passed in
- more more more stats to show
- subscription mecanism to send stats updates to be used for some ui, maybe
- model/bucket.go could proabably be moved inside the reporter package for now
