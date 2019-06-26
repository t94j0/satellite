# Options

The options example shows the basics of how to dynamically serve payloads based on an HTTP request.

In satellite, payloads with <name> will only be served when <name>.info is created. The file can be blank, but will eventually be populated with information about the endpoint.


## Setup

* Add the User-Agent for your browser to `useragent.info`. [This site][ua] will tell you what your User-Agent is, if you don't know how.

## Usage

1. Visit `/index.html`. The file is initially empty, so it will always be served.
2. Visit `/useragent`. This file will be served when the User-Agent string matches the one you entered in the [setup][#setup] step.
3. Visit `/exec`. This will fail since the string `abc123` is nowhere to be found in the request body.
4. Visit `/exec?abc123`. Now, the request will succeed since `abc123` exists in the request body.


## Exec

The `exec` option will run script specified by `exec.script` and will give the request body as stdin. If stdout matches the field `exec.output`, then the request will succeed.

[ua]: https://www.whatismybrowser.com/detect/what-is-my-user-agent
