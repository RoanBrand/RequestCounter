# RequestCounter
Services to handle and count requests.

# Run, Build and Development
- Requires Golang 1.18, Docker.
- Set required config in `.env`.
- See `Makefile` command to build and run.

## Cluster
- Single instance service.
- Counts the number of http requests made to it.
- Returns current count in a request.
- Basic async disk persistence.

## RequestCounter
- Multi instance service. Currently 3 replicas.
- Counts the number of http requests made to it.
- Makes request to cluster on behalf of client.
- Returns human readable informational message about node and cluster counts.

## nginx
- Client facing service. Publicy exposed.
- Reverse proxy to RequestCounter services.

### External libs used
- `github.com/pkg/errors`: useful for handling and bubbling up errors, with stack traces.
