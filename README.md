# terraform-ncloud-docs
## Overview
- This docs help to use terraform creation server


## Configuration (for run)
- Add configure file to call API
- Copy `account.yaml.sample` on root directory to `account.yaml` and fix it with your accessKey & secreyKey.
``` yaml
accounts:
- domain: "Pub"
  region: "KR"
  accessKey: {your access key}
  secretKey: {your secret key}
  apiUrl: "https://ncloud.apigw.ntruss.com"
- domain: "Fin"
  region: "FKR"
  accessKey: {your access key}
  secretKey: {your secret key}
  apiUrl: "https://fin-ncloud.apigw.fin-ntruss.com"
- domain: "Gov"
  region: "KR"
  accessKey: {your access key}
  secretKey: {your secret key}
  apiUrl: "https://ncloud.apigw.gov-ntruss.com"
```


## Contents

- [server_image_product](docs/server_image_product.md)