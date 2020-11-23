# Grafzahl
Small Docker Hub Rate Limits Exporter for Prometheus
## Options
### Flag
- ```--username=<username>```
- ```--password=<password>```
### Environment
- ```GRAFZAHL_PASSWORD```
- ```GRAFZAHL_USERNAME```
### Yaml
```yaml
username: <USERNAME>
password: <PASSWORD>
```
Path: ```/etc/grafzahl.yaml```/```~/grafzahl.yaml```
## Metrics
```test
 HELP docker_hub_rate_limit The maximal pulls for this account.
# TYPE docker_hub_rate_limit gauge
docker_hub_rate_limit 0
# HELP docker_hub_rate_limit_remaining The remaining pulls for this account.
# TYPE docker_hub_rate_limit_remaining gauge
docker_hub_rate_limit_remaining 0
```
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
