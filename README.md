# genaigallery

## Development

Easiest way to develop is using Podman (docker alternative). Install `podman` and `podman-compose` and run:

```bash
podman-compose up
```

By default it runs the app in port `5775`, but if that port is already used in your machine, create a file
named `.env` with contents:

```bash
WEB_HOST_PORT=5776  # or any other number
```

After starting the app it will be available at <http://localhost:5775> (adjust the port if necessary).

