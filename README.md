# prometheus-vmware-exporter
Collect metrics from an ESXi Host

## Build
```sh 
docker build -t prometheus-vmware-exporter .
```

## Run
```sh
docker run -b \
  --restart=always \
  --name=prometheus-vmware-exporter \
  --env=ESX_HOST esx.domain.local \
  --env=ESX_USERNAME user \
  --env=ESX_PASSWORD password \
  --env=ESX_LOG debug \
  prometheus-vmware-exporter 
```