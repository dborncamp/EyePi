# EyePi
A microservice that can count people entering a building.

## To run
Assuming you have go already set up...
It should be as simple as cloning this repo decending into the EyePie directory `cd EyePie` then running `dep ensure` followed by `./build_server.sh` followed by `${GOPATH}/bin/EyePi --config ${GOPATH}/src/github.com/dborncamp/EyePi/EyePi/settings.toml` on it.

You should then be able to go to http://localhost:8080/EyePi/services/hello?hello_text=ping and get a response and see the metrics at http://localhost:10001/metrics.