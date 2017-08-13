
# para

The utility allows me to process multiple commands concurrently instead of sequentially. This is particularly useful when scripting and piping commands together.

It takes commands from `stdin`. One command per line, and returns a JSON object in the following format:

```json
{
  "Results": [
    {
      "Command": "...",
      "Raw": "...",
      "Json": <JSON>
    },
    {
      "Command": "...",
      "Raw": "...",
      "Json": <JSON>
    }
  ]
}
```

Each input command will have an associated output entry in the `Results` property. Results order is not guaranteed.

## Properties

* `Command`: This is the command that was executed
* `Raw`: This is the raw consolidated output from running the command (stdout and stderr are combined)
* `Json`: JSON object representing the `Raw` output if successfully marshalled, `null` otherwise.

## TODO

* When running a lot of commands, it can freeze for a second depending on your workload, so it might be good to stagger the goroutines, or limit how many routines can run at a single time
* More/better error handling
* ...

