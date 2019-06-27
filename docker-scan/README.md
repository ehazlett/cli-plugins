# Docker Scan
Docker scan is an image vulnerability scanner and container risk analyzer.

```bash
Usage:  docker scan COMMAND

Docker Security Scanning

Commands:
  containers  Check containers for various risks
  images      Scan images for vulnerabilities using Aquasec Microscanner

Run 'docker scan COMMAND --help' for more information on a command.
```

# Image Scanning
Docker Scan uses the [Aquasec Microscanner](https://github.com/aquasecurity/microscanner) for vulnerability analysis.  You will need to register to get a token before you can use the scanner.  See [here](https://github.com/aquasecurity/microscanner#registering-for-a-token) for details.

```
Usage:  docker scan images

Scan images for vulnerabilities using Aquasec Microscanner

Options:
      --no-pull        disable image pulling
      --token string   microscanner token
```

Usage:

```bash
$> docker scan images --no-pull --token=$TOKEN ubuntu:xenial debian:latest redis:latest ubuntu:latest debian:jessie mariadb:latest
ID             TOTAL LOW MEDIUM HIGH MALWARE
ubuntu:xenial  11    8   3      0    0
ubuntu:latest  6     4   2      0    0
debian:jessie  0     0   0      0    0
debian:latest  0     0   0      0    0
redis:latest   0     0   0      0    0
mariadb:latest 8     6   2      0    0
```

# Container Scanning
Docker Scan can analyze containers for common, potential misconfigurations.  The following are currently checked:

- Privileged: are any containers running as privileged
- Published All Ports: do any containers have all ports published
- Host Mounts: are any containers bind mounting volumes from the host
- Cap Add: have any containers been launched with additonal capabilities

Usage

```bash
$> docker scan containers
ID            IMAGE              RUNNING  PRIVILEGED  PUBLISHEDALLPORTS  HOSTMOUNTS  CAPADD
b1378eb528be  docker-dev:carbon  false    true        false              2
40ca044ae373  alpine:latest      true     false       false              1           NET_ADMIN,MKNOD,KILL
29560cc8a55c  nginx:alpine       true     false       true               0
e292fa16707c  alpine:latest      true     false       false              1           NET_ADMIN
4a786667045f  docker-dev:carbon  false    true        false              2
ce07ba57c843  docker-dev:carbon  false    true        false              2
56cfc4864029  docker-dev:carbon  false    true        false              2
8b273f0d7337  docker-dev:carbon  false    true        false              2
93fca3e10864  docker-dev:carbon  false    true        false              2
ac8b6bf3c5e4  docker-dev:carbon  false    true        false              2
71952d6bf547  docker-dev:carbon  false    true        false              2
a467b8cce8bd  docker-dev:carbon  false    true        false              2
bcb3530d62dd  docker-dev:carbon  true     true        false              3
ab82a457bad4  docker-dev:carbon  false    true        false              2
cf88d432d1e1  docker-dev:carbon  false    true        false              2
d0ee0b012946  docker-dev:carbon  false    true        false              2
658c96db160f  docker-dev:carbon  false    true        false              2
eba18fdb8f59  docker-dev:carbon  true     true        false              2
6fa5ed25d5dd  docker-dev:carbon  false    true        false              2
0d74f15271ed  docker-dev:carbon  true     true        false              2
```
