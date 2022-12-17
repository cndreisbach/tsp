# tsp

`tsp` (Tailscale Proxy) is a reverse proxy that creates a virtual Tailscale machine for you.

## Example usage

```
TS_AUTHKEY=<your-auth-key> ./tsp -v -h testhost http://localhost:9000
```