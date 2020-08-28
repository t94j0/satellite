# satellite

Satellite is an web payload hosting service which filters requests to ensure the correct target is getting a payload. This can also be a useful service for hosting files that should be only accessed in very specific circumstances.


## Quickstart Guide

1. [Install satellite](https://github.com/t94j0/satellite/wiki/Installation) on Ubuntu using the .deb file

`dpkg -i satellite_X.X.X_linux_amd64.tar.gz`

2. Create file to serve

`echo '<h1>It worked!</h1>' > /var/www/html/index.html`

3. Create filtering file for index.html

`echo -e "authorized_useragents:\n- ayyylmao" > /var/www/html/index.html.info`

4. Run satellite

`systemctl start satellite`

5. Test satellite

This will return **It worked!**

`curl -k -A ayyylmao https://localhost/`

This will not

`curl -k https://localhost`


## Example Usage

To get hands-on experience with the options, check out the [examples](https://github.com/t94j0/satellite/tree/master/examples) folder. Replace your `server_root` with the sub-folder and try out the options.


## Wiki

For a more detailed explaination of how to use satellite, check out the [wiki](https://github.com/t94j0/satellite/wiki)


## Projects Used:

* [JA3 Server][ja3server]
* [MaxMind](https://www.maxmind.com/en/geoip2-databases)

[go]: https://golang.org/dl/
[ja3]: https://github.com/salesforce/ja3
[ja3server]: https://github.com/CapacitorSet/ja3-server
[server header]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Server
[contenttype]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
[issue]: https://golang.org/src/net/http/httputil/reverseproxy.go?s=3330:3391#307
