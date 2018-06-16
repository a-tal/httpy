"""Example python worker module."""


import time
import logging


LOG = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s:%(name)s:%(levelname)s: %(message)s",
)

REQS = {"count": None}


def go_init():
    """A (optional) init function, to be called during golang's init."""

    REQS["count"] = 0
    LOG.info('go_init called. REQS["count"] = %r', REQS["count"])

    # the format of the paths is up to you/your framework
    return {
        "/": ["GET"],
        "/ping": ["GET"],
        "/hello/:target": ["GET"],
    }


def go_request(method, path, params, query, headers, body):
    """Our per-request interface function.

    This function signature is absolute. You cannot deviate from it.

    Args:
        method: string HTTP method
        path: URL path (optional if patterned)
        params: dictionary {string: [string]} of path parameters
        query: dictionary of {string: [string]} of query string arguments
        headers: dictionary {string: [string]} of header values
        body: raw request body as string

    Returns:
        (status: int, body: string, headers: dict {string: [string]})
    """

    REQS["count"] += 1
    LOG.info(
        "golang request #%d received:\n  method: %s\n  path: %s\n  "
        "params: %r\n  query: %r\n  headers: %r\n  body: %s",
        REQS["count"],
        method,
        path,
        params,
        query,
        headers,
        body,
    )

    if "X-Sleep" in headers:
        time.sleep(float(headers["X-Sleep"][0]))

    return 200, "request #{:,d}".format(REQS["count"]), {"X-Test": ["1"]}
