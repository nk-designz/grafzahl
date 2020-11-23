# Grafzahl
Monitor your Dockerhub Rate-Limits in Prometheus on Port 6969! Nice!

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
## Run
### Docker
#### Step 1
Start the container
```bash
docker run --restart=always \
  -p 6969:6969 \
  -e "GRAFZAHL_PASSWORD=<PASSWORD>" \
  -e "GRAFZAHL_USERNAME=<USERNAME>" \
  me/grafzahl:latest
```
#### Step 2
Pet nearby cat.
### Native
#### Step 1
Move the binary into ```PATH```
```bash
sudo install grafzahl
```
#### Step 2
Add the config
```bash
vim /etc/grafzahl.yaml
```
#### Step 3
Run the process
```bash
./grafzahl &
```
_A service for systemd would be welcome_
#### Step 4 
Pet nearby cat nonetheless.
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
