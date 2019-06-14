# DEITYSHADOW


## Installation

TODO

## Configuration

DEITYSHADOW looks in `$HOME/.config/DEITYSHADOW/config.yml`, `$HOME/.DEITYSHADOW/config.yml`, and `/etc/DEITYSHADOW/config.yml` for service configuration.

| Key             | Configuration |
|:---------------:|:-------------:|
| server_root     | Server root directory |
| listen          | IP:port combination to listen on. IP can be removed to sigify listening on 0.0.0.0 |
| server_header   | Server header to give clients when requesting a page. More information can be found [here][server header] |
| management.ip   | IP (or range) which is allowed to view the management portal |
| management.path | Management server URL path. Takes precidence over any already-existing paths |
| ssl.key         | SSL key path |
| ssl.cert        | SSL cert path |


An example configuration is shown below.


```yaml
server_path: /var/www/html
listen: 127.0.0.1:8080

server_header: Apache/2.4.1 (Unix)

management:
  ip: 127.0.0.1
  path: /management

ssl:
  key: /home/<user>/.config/DEITYSHADOW/keys/key.unencrypted.pem
  cert: /home/<user>/.config/DEITYSHADOW/keys/cert.pem
```

## Serving Payloads

In the `server_root` directory chosen in the [configuration](#configuration) section, place any files you want to serve as a payload. Along with those files, create a `<payload_name>.info` file. For example, if you want to host a payload called `index.html`, make `index.html.info` as well. .info files are YML which contain filtering information for the payload you are hosting.

The only required field in a .info file is the `id` field. Every payload hosted must have a unique, greater-than-zero, integer value `id` field.

Here are all the filtering options for a file.

### id

ID of payload

#### Example

```yaml
id: 1
```

### serve

Number of times to serve file

#### Example

This will serve the file once before not allowing the file to be accessed anymore

```yaml
id: 1
serve: 1
```

### authorized_useragents

List of User Agent strings to allow

#### Example


```yaml
id: 1
authorized_useragents:
  - Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Mobile/15E148 Safari/604.1
  - Mozilla/5.0 (Windows Phone 10.0; Android 6.0.1; Microsoft; RM-1152) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Mobile Safari/537.36 Edge/15.15254
```

### authorized_iprange

List of IPs or IP ranges who are allowed to view a file

#### Example

```yaml
id: 1
authorized_iprange:
  - 192.168.0.1
  - 192.168.10.1/24
```

### authorized_methods

Authorized HTTP methods

#### Example

```yaml
id: 1
authorized_methods:
  - GET
  - PUT
```

### authorized_headers

Dictionary of headers which must be present

#### Example

```yaml
id: 1
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
id: 1
authorized_ja3:
  - e7d705a3286e19ea42f587b344ee6865
  - 6734f37431670b3ab4292b8f60f29984
```

### blacklist_iprange

Blacklisted IPs from accessing a payload

#### Example

```yaml
id: 1
blacklist_iprange:
  - 94.130.90.152
```

### prereq

Prerequisite IDs which must be hit, in order, before the payload will be served.

#### Example

In this case, when `/first` is requested, it is automatically served. When `/second` is accessed, the user will be served a 404 page. When `/first` is accessed, and then `/second` is accessed after, `/second` will be successfully hosted. When `/first` is accessed and then `/second` is accessed, you will finally be able to get `/payload`.

first.info

```yaml
id: 1
```

second.info

```yaml
id: 2
prereq:
  - 1
```

payload.info

```yaml
id: 3
prereq:
  - 1
  - 2
```

### content_type

Sets the Content-Type for the payload being served. More information about the Content-Type header can be found [here][contenttype]

#### Example

```yaml
id: 1
content_type: application/msword
```

### disposition

Sets Content-Disposition header for the payaload. There are two sub-keys: `type` and `file_name`. `type` can either be `inline` or `attachment`. `file_name` is the name for the attachment if the attachment option is chosen

#### Example


```yaml
id: 1
disposition:
  type: attachment
  file_name: file.docx
```


### exec

Executes a program, gives the HTTP request to stdin, and checks stdout against an output variable.

#### Example

```yaml
id: 1
exec:
  script: /home/user/test.py
  output: success
```


### add_headers

Adds the header to all HTTP responses.

#### Example

```yaml
id: 1
add_headers:
  Accept-Encoding: gzip
```

### add_headers_success

Adds the header to an HTTP response if the page was successfully reached


### add_headers_failure

Adds the header to an HTTP response if the request was denied


### times_served

Number of times a payload has been accessed. This variable is for DEITYSHADOW to do record-keeping.


### not_serving

Boolean to determine if the file should be served. This is mostly used by the DEITYSHADOW server for record-keeping, but can be set manually to now allow a payload to be hosted.

#### Example

```yaml
id: 1
not_serving: true
```

## Management API

The management API can be reached on the path and with the IPs specific in the [service configuration](#configuration).

GET requests will return a list of paths and information about those paths.

POST requests can have the following options:

| Key | Description |
|-----|-------------|
| id  | ID of path to request |
| reset | boolean. When `true`, the ID is reset to `times_served` = 0 and `not_serving` = false |

Example JSON body if one wished to reset ID 1.

```json
{ "id": 1, "reset": true }
```


## TODO

* Management API ideas
* Systemd service file


Open Source Projects Used:

* [JA3 Server][ja3server]

[ja3]: https://github.com/salesforce/ja3
[ja3server]: https://github.com/CapacitorSet/ja3-server
[server header]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Server
[contenttype]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
