
# para

The utility allows me to process multiple commands concurrently instead of sequentially. This is particularly useful when scripting and piping commands together.

It takes commands from `stdin`. One command per line, and returns a JSON object in the following format:

```json
{
  "Results": [
    {
      "Command": "...",
      "Raw": "...",
      "Json": <JSON>,
      "ExecutionTime": "5s"
    },
    {
      "Command": "...",
      "Raw": "...",
      "Json": <JSON>,
      "ExecutionTime": "10ms"
    }
  ]
}
```

Each input command will have an associated output entry in the `Results` property. Results order is not guaranteed.

## Properties

* `Command`: This is the command that was executed
* `Raw`: This is the raw consolidated output from running the command (stdout and stderr are combined)
* `Json`: JSON object representing the `Raw` output if successfully marshalled, `null` otherwise.
* `ExecutionTime`: A string representing the [duration](https://golang.org/pkg/time/#Duration.String) of the command.

## Usage

```
Usage of para:
  -c int
        Maximum number of commands to run at the same time (default 10)
  -file string
        Path to commands file
  -pp
        Pretty print json output
```

## Example

### Top SQS Queues

Using [jq](https://stedolan.github.io/jq) to parse JSON outputs:

```
$ aws sqs list-queues --queue-name-prefix production \
  | jq -r '.QueueUrls | join("\n")' \
  | while read url; do echo "aws sqs get-queue-attributes --queue-url $url --attribute-names QueueArn ApproximateNumberOfMessages"; done \
  | para \
  | jq -r '.Results[].Json.Attributes | .ApproximateNumberOfMessages +"\t"+ .QueueArn' \
  | sort -rn
```

Without `para` I would have to fetch each queue attributes sequentially, it takes over 2 minutes to fetch 150. Using `para` this drops to 15 seconds!

## TODO

* When running a lot of commands, it can freeze for a second depending on your workload, so it might be good to stagger the goroutines, or limit how many routines can run at a single time
* More/better error handling
* ...

