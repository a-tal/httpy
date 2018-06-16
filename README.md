# HTTPy

Ever wanted to embed some python in your golang http app? Well now you can!

Dead simple interfacing between golang and python. Implementation of route
parsing is left to the developer/framework's choice.


## Features

- asynchronous embedded python in your golang app
- a memory-leak free interface to CPython
- flexibility in the dead simple interfacing to & from python


## Pictures tell 1,000 words

Check out the example directory for a working example. Run the example locally
with the Dockerfile in the root of this repo.


## Paths and Params

You should note that the params being passed from golang -> python are
not complete in the example. If you were using httprouter, this would
interface directly with their third handler argument. If however, you
prefer to parse the parameters in python, then by all means do that.

Either path and/or parameters _should_ be passed to python, probably.
Or maybe your python app doesn't need either of those (single service),
you do you!


## Embedded Python

Currently only Python3.6 is supported. Other 3.x releases could probably be
added without too much effort. There are no plans to add support for Python2.

Note that the function signatures of `go_init` and `go_request` defined in the
`example/worker.py` module must be re-implemented exactly. Failure to do so
will probably result in a segfault.


### Python Request Interface

Write your python interface for `httpy.Request` with the following signature:

```python
def your_request_interface(method, path, params, query, headers, body):
```

Where `method`, `path` and `body` are strings, and `params`, `query` and
`headers` are dictionaries with string keys and a list of strings for values.

This interface must return three values:

- an integer status code.
- the request body as a string
- response headers as a dictionary of string keys to list of string values


### Python Init Interface

If you have dynamic routes, or otherwise want to initialize some things in
python on app init, define a function somewhere with the following signature:

```python
def your_init_function():
```

Your function should return a dictionary of string to list of strings, which
can then be used to initialize the mux in your golang router. The return format
should be: `{"route": ["method", ...]}` where route is a string, formatted as
per whatever http framework you're using, and the supported methods are per
route as a list of strings.

Interesting to note here though, this isn't enforced or used anywhere in
`httpy`. You are free to use this init return however you like, as long as
the type remains the same. ie; if you wanted to arrange your data into:
`{"method": ["route", ...]}` you are completely free to do so.

Also note an initialization interface is entirely optional. However, calling
`httpy.Init` is **not**. It must be called (only once) before `httpy.Request`
is used. To skip the python init function, call `httpy.Init` with empty strings
for `initModule` and `initFunction`.


## Performance

It's not great yet, but it's also not terrible? Here are some local results:

```bash
$ wrk -t12 -c400 -d30s --timeout=10 --latency http://localhost:8080/python
Running 30s test @ http://localhost:8080/python
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   141.23ms   99.61ms   1.19s    65.36%
    Req/Sec   218.84     70.01   797.00     79.90%
  Latency Distribution
     50%  149.69ms
     75%  161.28ms
     90%  292.35ms
     99%  462.41ms
  77955 requests in 30.10s, 10.62MB read
  Socket errors: connect 0, read 617, write 1, timeout 0
Requests/sec:   2589.66
Transfer/sec:    361.26KB
```

The next major increase in performance will come with implementing a C event
loop instead of using golang's `runtime.LockOSThread`.

Feel free to build the example and compare for yourself. The socket errors are
just docker being docker, happens to `/golang` as well. Performance would
probably also increase a bunch by running it in not-docker.

For comparison, benchmarking the same host without using python:

```bash
$ wrk -t12 -c400 -d30s --timeout=10 --latency http://localhost:8080/golang
Running 30s test @ http://localhost:8080/golang
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    52.02ms  126.73ms   1.07s    95.21%
    Req/Sec     1.12k   291.00     2.81k    80.70%
  Latency Distribution
     50%   28.49ms
     75%   33.53ms
     90%   38.40ms
     99%  792.79ms
  382759 requests in 30.09s, 43.07MB read
  Socket errors: connect 0, read 792, write 0, timeout 0
Requests/sec:  12719.90
Transfer/sec:      1.43MB
```

But I mean, if you were *really* interested in performance, you'd use rust.
Maybe I should work on doing this in a crate and compare the two...

:grin:
