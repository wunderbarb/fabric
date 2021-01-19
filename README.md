This is only for test purpose.

The configuration files expect that the "root" directory of the project is `Dev/github.com/blockchain1`.  Would it be different, then `config.toml`, `config.yaml` and `connection-org1.yaml` need to be updated correspondingly.

The remote blockchain is operational running an instance of fabcar.  The remote blockchain can be accessed from any IP address.

`go test` fails.  The same commands using a traditional access work perfectly.

The status of `DISCOVERY_AS_LOCAL_HOST` is in the log file `bc.log`.

The directory `logs` hold four logs from the remote peer0 in DEBUG mode. The logs with prefix `normal_` are for transaction using the standard mode whereas the logs without the prefix are for transaction using the gateway.