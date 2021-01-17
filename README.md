# prometheus-vmware-exporter
Collect metrics ESXi Host

## Build
```sh 
docker build -t prometheus-vmware-exporter .
```

## Run
```sh
docker run -d \
  --restart=always \
  --name=prometheus-vmware-exporter \
  --env=ESX_HOST esx.domain.local \
  --env=ESX_USERNAME user \
  --env=ESX_PASSWORD password \
  --env=ESX_LOG debug \
  prometheus-vmware-exporter 
```
