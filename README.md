# disys-handin-5
GoLang solution for ITU course Distributed System 2022 - hand-in 5

## How to run
First, run the servers:
```
go run server/server.go --port <port> --duration <duration>
```
Three server should be run, with the ports 5000, 5001 and 5002 as this is what the client will look for.

The duration is input as a number and a unit, e.g. `1m`. Defaults to 1 minute.
This is the duration the auction will run for before declaring a winner


Then, run the client:
`go run client/client.go`

### Notes
The servers will write logs to a file named `server_<port>.log`.
The client will write to `client_<id>.log`
