# Proxy

The proxy example shows how to use satellite when deploying a Cobalt Strike beacon through a redirector. More information about Cobalt Strike and redirectors can be found [here][blog post].

If you put a proxy.yml file in your server root, satellite will proxy URIs requests to the target host.


## Setup

1. Set up a Cobalt Strike teamserver with the [Amazon malleable profile][profile]
2. Create an HTTPS listener
3. Add your host to the `proxy.yml` file
4. Create a Stageless beacon executable and replace `beacon.exe`

## Usage

1. In a Windows VM, download beacon.exe (It will only be served once)
2. Execute beacon.exe
3. Verify the proxy works by interacting with beacon in Cobalt Strike


[blog post]: https://blog.cobaltstrike.com/2014/01/14/cloud-based-redirectors-for-distributed-hacking/
[profile]: https://github.com/rsmudge/Malleable-C2-Profiles/blob/master/normal/amazon.profile
