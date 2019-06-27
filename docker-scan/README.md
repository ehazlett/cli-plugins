# Docker Scan
Docker scan is an image vulnerability scanner and container risk analyzer.

Note: be careful when scanning a large amount of images at once.  You might get [rate limited](https://github.com/aquasecurity/microscanner#fair-use-policy).

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
$> docker scan images --no-pull --token=$TOKEN ubuntu:latest ubuntu:xenial debian:latest redis:latest debian:jessie mariadb:latest postgres:latest centos:latest
ID              TOTAL LOW MEDIUM HIGH MALWARE
ubuntu:latest   6     4   2      0    0
ubuntu:xenial   11    8   3      0    0
debian:latest   0     0   0      0    0
redis:latest    0     0   0      0    0
debian:jessie   0     0   0      0    0
mariadb:latest  8     6   2      0    0
postgres:latest 1     0   0      1    0
centos:latest   52    19  20     13   0
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
ID            IMAGE			RUNNING  PRIVILEGED  PUBLISHEDALLPORTS  HOSTMOUNTS  CAPADD
b1378eb528be  ehazlett/demo:latest	false    true        false              2
40ca044ae373  alpine:latest		true     false       false              1           NET_ADMIN,MKNOD,KILL
29560cc8a55c  nginx:alpine		true     false       true               0
e292fa16707c  alpine:latest		true     false       false              1           NET_ADMIN
```
