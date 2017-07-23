## Advent of Code, Day 10

### Comments

* Whilst all in one file, certain functions like getInputLines would
be easy to extract out. Hardcoding filename was done for expediency.
* Using goroutine and channel in getInputLines() was done to duplicate
generator behaviour for (imo) clearer usage of getInputLines in parseInputData()
* parseInputData was made concurrent as ordering of instruction lines is
unimportant as well as each instruction line being usable independently
from the next and thus lends itself nicely to concurrent running.
* I chose to affix some logic to the Bots and some in the Process loop. Separated
through which behaviours are bot responsibilities and which behaviours require
interaction with other bins/bots.
* Another possible location where I could implement concurrency is in the main
Process thread. However this would have required orchestration between the botMap
and clones of bots. The complexity of this orchestration could well negate any
 positive effects of adding concurrency at this point.

### Requirements
* Golang (1.83 used)
* Git

### Usage

Download from github.com, suggested command

```go get github.com/kdvy/advent10```

Compile

```go build```

Run

```./neuliontest```

#### Comments from code for convenience
 

ParseInputData converts the parsed text into Bot state. Written to return a state diff such that the changes
can be applied later. This prevents shared access errors reading from the botMap whilst running concurrently.

Process recursively checks and executes the bots actions if they are in the ready state to do so. The recursion
runs the same checks against any modified bots.

Use a goroutine to read from testdata file so that we can
iterate over the lines. Rather than wait for the whole file to be in memory
or embed parsing functions into getInputLines.