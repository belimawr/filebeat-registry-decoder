# filebeat-registry-decoder
Filebeat registry decoder reads Filebeat registry files (checkpoint,
log or a mix of both) and prints the JSON entries in a more
human-readable format. Currently it converts the `updated` field from
an integer array to a human-readable timestamp.

The `example.ndjson` file contains two entries, one in the format used
by the registry checkpoint and another in the format used by the
registry log.

## Usage
1. Build it with to build: `go build -o filebeat-registry-decoder  .`
2. Run it on a ndjson file: `./filebeat-registry-decoder example.ndjson`

## Example
Given those two entries
```
{"_key":"filestream::my-log-file::native::1041043-66305","meta":{"source":"/tmp/foo.log","identifier_name":"native"},"ttl":1800000000000,"updated":[502979742,1696416120],"cursor":{"offset":458}}
{"k":"filestream::my-log-file::native::1040242-66305","v":{"ttl":1800000000000,"updated":[507500968,1701686520],"cursor":{"offset":914},"meta":{"source":"/tmp/bar.log","identifier_name":"native"}}}

```

Using `filebeat-registry-decoder` and `jq` we can make the entries
human readable and sort the keys:
```
$ ./filebeat-registry-decoder example.ndjson| jq -c -M --sort-keys 
{"_key":"filestream::my-log-file::native::1041043-66305","cursor":{"offset":458},"meta":{"identifier_name":"native","source":"/tmp/foo.log"},"ttl":1800000000000,"updated":"2023-10-04T12:42:00.502979742+02:00"}
{"k":"filestream::my-log-file::native::1040242-66305","v":{"cursor":{"offset":914},"meta":{"identifier_name":"native","source":"/tmp/bar.log"},"ttl":1800000000000,"updated":"2023-12-04T11:42:00.507500968+01:00"}}
```
