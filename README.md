# Traefik Block Regex Urls

A [Traefik](https://github.com/containous/traefik) plugin to block access to certain paths using a list of regex values and return a defined status code.

## Configuration

Sample configuration in Traefik.

### Configuration as local plugin

traefik.yml

```yaml
log:
  level: INFO
experimental:
  localPlugins:
    block-regex-urls:
      moduleName: github.com/shantanugadgil/traefik-block-regex-urls
```

dynamic-configuration.yml

```yaml
http:
  middlewares:
    block-scan-paths:
      plugin:
        block-regex-urls:
          allowLocalRequests: true
          regex:
            - "^something.mydomain.tld/scan?uid=12345(.*)&gid=6789(.*)"
            - "^something.mydomain.tld/scan?uid=345$"
          statusCode: 404
```

docker-compose.yml

```yaml
services:
  traefik:
    image: traefik
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /docker/config/traefik/traefik.yml:/etc/traefik/traefik.yml
      - /docker/config/traefik/dynamic-configuration.yml:/etc/traefik/dynamic-configuration.yml
      - /plugin/traefik-block-regex-urls:/plugins-local/src/github.com/shantanugadgil/traefik-block-regex-urls/
    ports:
      - "80:80"
  hello:
    image: containous/whoami
    labels:
      - traefik.enable=true
      - traefik.http.routers.hello.entrypoints=http
      - traefik.http.routers.hello.rule=Host(`hello.localhost`)
      - traefik.http.services.hello.loadbalancer.server.port=80
      - traefik.http.routers.hello.middlewares=my-plugin@file
```

## Sample configuration

- `allowLocalRequests`: If set to true, will not block request from [Private IP Ranges](https://de.wikipedia.org/wiki/Private_IP-Adresse)
- `regex`:  List of regex values to use for url blocking.
- `statusCode`: Return value of the status code.

```yaml
my-block-regex-urls:
  plugin:
    block-regex-urls:
      allowLocalRequests: true
      regex:
        - "^something.mydomain.tld/scan?uid=12345(.*)&gid=6789(.*)"
        - "^something.mydomain.tld/scan?uid=345$"
      statusCode: 418
```

## Contributors

| [<img alt="JonasSchubert" src="https://secure.gravatar.com/avatar/835215bfb654d58acb595c64f107d052?s=180&d=identicon" width="117"/>](https://code.schubert.zone/jonas-schubert) |
| :---------------------------------------------------------------------------------------------------------------------------------------: |
| [Jonas Schubert](https://code.schubert.zone/jonas-schubert) |

## License

traefik-block-regex-urls is distributed under the MIT license. [See LICENSE](LICENSE) for details.

```
MIT License

Copyright (c) 2023-today Jonas Schubert

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
