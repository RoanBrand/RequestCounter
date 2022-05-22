# RequestCounter
Services to handle and count requests.

# Run, Build and Development
- Requires Golang 1.18, Docker.
- See `Makefile` useful commands.

## Cluster
- Single instance service.
- Counts the number of http requests made to it.
- Returns current count in a request.

## RequestCounter
- Client facing service.
- Counts the number of http requests made to it.
- Makes request to cluster on behalf of client.
- Returns human readable informational message about node and cluster counts.
