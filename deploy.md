# Deloy to Fly.io

```console
SERVICE_NAME=bk-srv
fly auth login
fly create "${SERVICE_NAME}"
fly volumes create bkdata -a "${SERVICE_NAME}" --size 1
fly deploy -a "${SERVICE_NAME}"
```
