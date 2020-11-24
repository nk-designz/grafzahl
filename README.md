# Grafzahl
Monitor your Dockerhub Rate-Limits in Prometheus on Port 6969! Nice!

## Options
| Flag | Environment | Yaml |
| --- | --- | --- |
| ```--username=<username>``` | ```GRAFZAHL_USERNAME``` | ```username: <USERNAME>``` |
| ```--password=<password>``` | ```GRAFZAHL_PASSWORD``` | ```password: <PASSWORD>``` |

Path: ```/etc/grafzahl.yaml``` or ```~/grafzahl.yaml```

## Metrics
```ebnf
 HELP docker_hub_rate_limit The maximal pulls for this account.
# TYPE docker_hub_rate_limit gauge
docker_hub_rate_limit 0
# HELP docker_hub_rate_limit_remaining The remaining pulls for this account.
# TYPE docker_hub_rate_limit_remaining gauge
docker_hub_rate_limit_remaining 0
```
[metrics](http://localhost:9696/metrics)
## Run
### Docker
1. Start the container
   ```bash
   docker run --restart=always \
     -p 6969:6969 \
     -e "GRAFZAHL_PASSWORD=<PASSWORD>" \
     -e "GRAFZAHL_USERNAME=<USERNAME>" \
     me/grafzahl:latest
   ```
2. Pet nearby cat.
### Native
1. Move the binary into ```PATH```
   ```bash
   sudo install grafzahl
   ```
2. Add the config
   ```bash
   vim /etc/grafzahl.yaml
   ```
3. Run the process
   ```bash
   ./grafzahl &
   ```
   _A service for systemd would be welcome_
4. Pet nearby cat nonetheless.
## Build
### Go
```bash
go get -v
go build .
```
### Docker
```bash
docker build . -t <user>/grafzahl:<version>
docker push <user>/grafzahl:<version>
```
