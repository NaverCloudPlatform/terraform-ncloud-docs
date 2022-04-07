# terraform-ncloud-docs
## Overview
- This docs help to use terraform creation server

## Configuration (for run)
- Add configure file to call API
- or Refer [Set CLI API Authentication key](https://cli.ncloud-docs.com/docs/guide-userguide) > Execute ncloud configure
```go
$ cat $HOME/.ncloud/configure
[DEFAULT]
ncloud_access_key_id = YOUR_ACCESS_ID
ncloud_secret_access_key = YOUR_SECRET_ACCESS_KEY
ncloud_api_url = https://ncloud.apigw.ntruss.com
```

## Contents

- [server_image_product](docs/server_image_product.md)