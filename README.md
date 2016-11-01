# calldedup

> Go function call de-duping

Reuse results of a function for calls received from other goroutines while the
function was still executing

Useful for saving resources during operations of which the results would be
nearly identical for overlapping calls

Example uses:

* Database read queries
* Output of shell commands
* HTTP GETs
