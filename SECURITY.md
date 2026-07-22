# Security

## Supported version

Only the most recent tagged version is supported with security fixes.

## Deployment boundary

Disk Insight has no built-in authentication and is intended for a trusted local machine or private LAN. Do not expose it directly to the public internet. Use an authenticated TLS reverse proxy when remote access is required.

The recommended container configuration:

- mounts analyzed directories read-only;
- keeps the container root filesystem read-only;
- drops all Linux capabilities;
- enables `no-new-privileges`;
- runs as UID/GID `10001`;
- never mounts the Docker socket;
- exposes metadata and relative paths only, with no file-content or download endpoint.

## Reporting a vulnerability

Please use GitHub's private vulnerability reporting feature for this repository. Do not include credentials, private filenames, private directory structures, or other sensitive scan data in a public issue.
