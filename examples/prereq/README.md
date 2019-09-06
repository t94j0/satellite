# Prereq

The prereq example shows how the `prereq` option in satellite will allow payloads to be served after other paths have been successfully requested.

The prereq option takes a list of paths which must be successfully requested by an IP before the target request will be completed successfully. In this example, `/first`, `/second`, and `/third` must be requested in order before `/payload` will be served.

## Usage

1. Request `/payload` and notice that you will fail to request it
2. Request `/first`, then `/second`, then `/third` in order
3. Request `/payload` again. This time, it should succeed.
