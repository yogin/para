
# para

The utility allows me to process multiple commands concurrently instead of sequentially. This is particularly useful when scripting and piping commands together.

It takes commands from `stdin`. One command per line, and returns a JSON string in the following format:

```json
[{
  "Command": "...",
  "Raw": "...",
  "Json": <JSON>
},{
  "Command": "...",
  "Raw": "...",
  "Json": <JSON>
}]
```

Each input command will have an associated output entry. Results order is not guaranteed.

## Properties

* `Command`: This is the command that was executed
* `Raw`: This is the raw consolidated output from running the command (stdout and stderr are combined)
* `Json`: JSON object representing the `Raw` output if successfully marshalled, `null` otherwise.

## Example

If I want to check the state (number of messages) of my queues in SQS, without `para` I would do something like this:

```
aws sqs list-queues --queue-name-prefix production \ 
  | jsawk 'return this.QueueUrls.join("\n")' \
  | while read url; do aws sqs get-queue-attributes --queue-url $url --attribute-names QueueArn ApproximateNumberOfMessages \
  | jsawk 'return out(this.Attributes.ApproximateNumberOfMessages+"\t"+this.Attributes.QueueArn)'; done \
  | sort -rn
```

And this works fine, but if you have N queues, this will run N+1 API calls sequentially. In my case I had 158 queues, and running all these calls would take about 2 minutes and 10 seconds.

Using `para` I can now get the same result in 16 seconds...

```
aws sqs list-queues --queue-name-prefix production \
  | jsawk 'return this.QueueUrls.join("\n")' \
  | while read url; do echo "aws sqs get-queue-attributes --queue-url $url --attribute-names QueueArn ApproximateNumberOfMessages"; done \
  | para \
  | jsawk -n 'return out(this.Json.Attributes.ApproximateNumberOfMessages+"\t"+this.Json.Attributes.QueueArn)' \
  | sort -rn
```

