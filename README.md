# satellite

Satellite is an web payload hosting service which filters requests to ensure the correct target is getting a payload. This can also be a useful service for hosting files that should be only accessed in very specific circumstances.


## Installation

1. Create SSL keys

`openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365`
`openssl rsa -in key.pem -out key.unencrypted.pem -passin pass:<pass>`

2. Build Project - Must have go environment set up

`go get -u github.com/t94j0/satellite`

3. Run Project

`satellite`

## Configuration

Satellite looks in `$HOME/.config/satellite/config.yml`, `$HOME/.satellite/config.yml`, and `/etc/satellite/config.yml` for service configuration.

| Key                | Configuration |
|:-------------------|:--------------|
| server_root        | Server root directory |
| listen             | IP:port combination to listen on. If this IP is removed, the service will listen on 0.0.0.0 |
| server_header      | Server header to give clients when requesting a page. More information can be found [here][server header] |
| management.ip      | IP (or range) which is allowed to view the management portal |
| management.path    | Management server URL path. If this option is not specified, the route does not get created. Takes precedence over any already-existing paths |
| ssl.key            | SSL key path |
| ssl.cert           | SSL cert path |
| index              | Index path for when a user requests `/` |
| not_found.redirect | Redirect to give page when a page isn't found |
| not_found.render   | Path of file to render when a page isn't found |
| log_level          | Minimum log level. Options are `panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace` |


An example configuration is shown below.


```yaml
server_path: /var/www/html
listen: 127.0.0.1:8080
index: /index.html

not_found:
  redirect: https://amazon.com

server_header: Apache/2.4.1 (Unix)

management:
  ip: 127.0.0.1
  path: /management

ssl:
  key: /home/<user>/.config/satellite/keys/key.unencrypted.pem
  cert: /home/<user>/.config/satellite/keys/cert.pem
```


## Serving Payloads

In the `server_root` directory chosen in the [configuration](#configuration) section, place any files you want to serve as a payload. Along with those files, create a `<payload_name>.info` file. For example, if you want to host a payload called `index.html`, make `index.html.info` as well. .info files are YML which contain filtering information for the payload you are hosting.

Here are all the filtering options for a file.


### serve

Number of times to serve file

#### Example

This will serve the file once before not allowing the file to be accessed anymore

```yaml
serve: 1
```

### authorized_useragents

List of User Agent strings to allow

#### Example


```yaml
authorized_useragents:
  - Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Mobile/15E148 Safari/604.1
  - Mozilla/5.0 (Windows Phone 10.0; Android 6.0.1; Microsoft; RM-1152) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Mobile Safari/537.36 Edge/15.15254
```

### authorized_iprange

List of IPs or IP ranges who are allowed to view a file

#### Example

```yaml
authorized_iprange:
  - 192.168.0.1
  - 192.168.10.1/24
```

### authorized_methods

Authorized HTTP methods

#### Example

```yaml
authorized_methods:
  - GET
  - PUT
```

### authorized_headers

Dictionary of headers which must be present

#### Example

```yaml
authorized_headers:
  Hacked: yes
```

The request with a header of `Hacked` and a value of `yes`, like below, would be allowed to access the payload.

```yaml
GET / HTTP/1.1
Host: google.com
Hacked: yes
...
```

### authorized_ja3

Authorized JA3 hashes to access the file. More information about JA3 can be found [here][ja3].

#### Example

```yaml
authorized_ja3:
  - e7d705a3286e19ea42f587b344ee6865
  - 6734f37431670b3ab4292b8f60f29984
```

### blacklist_iprange

Blacklisted IPs from accessing a payload

#### Example

```yaml
blacklist_iprange:
  - 94.130.90.152
```

### prereq

Prerequisite paths which must be hit, in order, before the payload will be served.

#### Example

In this case, when `/first` is requested, it is automatically served. When `/second` is accessed, the user will be served a 404 page. When `/first` is accessed, and then `/second` is accessed after, `/second` will be successfully hosted. When `/first` is accessed and then `/second` is accessed, you will finally be able to get `/payload`.

first.info

```yaml
```

second.info

```yaml
prereq:
  - /index
```

payload.info

```yaml
prereq:
  - /first
  - /second
```

### content_type

Sets the Content-Type for the payload being served. More information about the Content-Type header can be found [here][contenttype]

#### Example

```yaml
content_type: application/msword
```

### disposition

Sets Content-Disposition header for the payload. There are two sub-keys: `type` and `file_name`. `type` can either be `inline` or `attachment`. `file_name` is the name for the attachment if the attachment option is chosen

#### Example


```yaml
disposition:
  type: attachment
  file_name: file.docx
```


### exec

Executes a program, gives the HTTP request to stdin, and checks stdout against an output variable.

#### Example

```yaml
exec:
  script: /home/user/test.py
  output: success
```


### add_headers

Adds the header to all HTTP responses.

#### Example

```yaml
add_headers:
  Accept-Encoding: gzip
```

### add_headers_success

Adds the header to an HTTP response if the page was successfully reached


### add_headers_failure

Adds the header to an HTTP response if the request was denied


### times_served

Number of times a payload has been accessed. This variable is for satellite to do record-keeping.


### not_serving

Boolean to determine if the file should be served. This is mostly used by the satellite server for record-keeping, but can be set manually to now allow a payload to be hosted.

#### Example

```yaml
not_serving: true
```


### on_failure

Specifies what happens when the request does not match the prerequisites. There are two options: redirection, available through `on_failure.redirect`, and rendering a web page, available through `on_failure.render`.

#### Example: Redirection

```yaml
on_failure:
  redirect: https://google.com
```

#### Example: Rendering

```yaml
on_failure:
  redirect: /index.html
```


### proxy

Proxy route to a different address

#### Example

```yaml
proxy: http://localhost:2222
```



## Management API

The management API can be reached on the path and with the IPs specific in the [service configuration](#configuration). For these examples, I will assume the operator is using `/management` as the management path. Use JSON format for all POST requests.

GET /management - List all IDs and information about those IDs

POST /management/reset - Reset `times_served` and `not_serving` for target path

| Key  | Description      |
|------|------------------|
| path | Path to reset    |

```json
{ "path": "/index.html" }
```

POST /management/new - Create new path with data. Warning: Users can arbitrarily write to the filesystem by using directory traversal on path.Path.

| Key | Description |
|-----|-------------|
| path | Info object to upload. Look at the options in `paths.go` |
| file | String with file data as base64 |

```json
{
  "path": {
    "Path": "/info/test.html",
    "Serve": 1,
  },
  "file": "PGI+SGVsbG8hPC9iPgo="
}
```

## TODO

* Specific redirection per IP
* Timeout on ClientID
* `handler` package returns an http.Handler object instead of a full HTTP server


## Open Source Projects Used:

* [JA3 Server][ja3server]

[ja3]: https://github.com/salesforce/ja3
[ja3server]: https://github.com/CapacitorSet/ja3-server
[server header]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Server
[contenttype]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
