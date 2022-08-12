# Load-balancer using Go (Least connections)

### Run
    docker-compose up

### Description
Program brings a list of urls from a `config.json` file and serves these services with a separate load-balancing service in localhost.

`spam.go` file is executing to simulate the requests to a load-balancing service.

To visualize the balancer work, the 1st service in a list shuts down after 100s of execution and runs again on the 200th second.

Passive Health Check runs every 45s.

Logs prints into the console.
