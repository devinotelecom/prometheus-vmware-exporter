# prometheus-vmware-exporter
Collect metrics ESXi Host

## Build
```sh 
docker build -t prometheus-vmware-exporter .
```

## Run
Fill in `ESXI_ADDRESS`, `USER`, and `PASSWORD` with your information.

```sh
docker run -d \
  -p="9512:9512" \
  --restart=always \
  --name=prometheus-vmware-exporter \
  --env="ESX_HOST=<ESXI_ADDRESS>" \
  --env="ESX_USERNAME=<USER>" \
  --env="ESX_PASSWORD=<PASSWORD>" \
  --env="ESX_LOG=debug" \
  prometheus-vmware-exporter 
```
