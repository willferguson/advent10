package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"strconv"
)

var botMap map[int]*bot
var binMap map[int]*bin

type targetType int
type diffType int

const (
	TYPE_BOT targetType = 1
	TYPE_BIN targetType = 2
)

const (
	TYPE_VAL diffType = 1
	TYPE_TARGET diffType =2
)

type bot struct {
	Id         int
	Values     [2]int // 0 - low|default, 1 - high
	HighTargetId int
	HighTargetType targetType
	LowTargetId  int
	LowTargetType targetType
}

type botDiff struct {
	BotId          int
	Type           diffType
	Value          int
	HighTargetId   int
	HighTargetType targetType
	LowTargetId    int
	LowTargetType  targetType
}

func (d botDiff) Apply(b *bot) {
	if (d.Type == TYPE_TARGET) {
		b.LowTargetType = d.LowTargetType
		b.LowTargetId = d.LowTargetId
		b.HighTargetType = d.HighTargetType
		b.HighTargetId = d.HighTargetId
	} else {
		b.Receive(d.Value)
	}
}

func (b *bot) Receive(v int) {
	if v > b.Values[1] {
		b.Values[0] = b.Values[1]
		b.Values[1] = v
	} else if v > b.Values[0] {
		b.Values[0] = v
	}
}

func (b *bot) Clear() {
	b.Values = [2]int{0, 0}
}

func (b *bot) ReadyToSend() bool {
	if (b.Values[0] != 0 && b.Values[1] !=0 && b.LowTargetType != 0 && b.HighTargetType != 0) {
		low := "bin "
		if b.LowTargetType == TYPE_BOT {
			low = "bot "
		}
		low = low + strconv.Itoa(b.LowTargetId)

		high := "bin "
		if b.HighTargetType == TYPE_BOT {
			high = "bot "
		}
		high = high + strconv.Itoa(b.HighTargetId)
		fmt.Printf("Bot %d will give low: %d to %s high: %d to %s \n", b.Id, b.Values[0], low, b.Values[1], high)
		return true
	}
	return false
}

func (b *bot) SetHighTarget(id int, t targetType) {
	b.HighTargetId = id
	b.HighTargetType = t
}

func (b *bot) SetLowTarget(id int, t targetType) {
	b.LowTargetId = id
	b.LowTargetType = t
}

type bin struct {
	Id    int
	Value int
}

func (b *bin) Receive(v int) {
	b.Value = v
}

func main() {
	botMap = make(map[int]*bot)
	binMap = make(map[int]*bin)
	for d := range parseInputData() {
		b := GetBot(d.BotId)
		d.Apply(b)
		Process(b.Id)
	}
}

/**
ParseInputData converts the parsed text into Bot state. Written to return a state diff such that the changes
can be applied later. This prevents shared access errors reading from the botMap whilst running concurrently.
 */
func parseInputData() <-chan botDiff {
	ch := make(chan botDiff, 10)
	go func() {
		defer close(ch)
		for line := range getInputLines() {
			if line == "" {
				break
			}

			if line[0] == 'v' {
				var v, b int
				_, err := fmt.Sscanf(line, "value %d goes to bot %d", &v, &b)
				if err != nil {
					log.Fatal(err.Error())
				}
				d := botDiff{BotId: b, Value: v, Type: TYPE_VAL}
				ch<- d
			} else if line[0] == 'b' {
				var b, hid, lid int
				var ht, lt string
				_, err := fmt.Sscanf(line, "bot %d gives low to %s %d and high to %s %d", &b, &lt, &lid, &ht, &hid)
				if err != nil {
					log.Fatal(err.Error())
				}
				d := botDiff{BotId: b, Type: TYPE_TARGET}
				d.HighTargetId = hid
				d.HighTargetType = TYPE_BOT
				if ht == "output" {
					d.HighTargetType = TYPE_BIN
				}
				d.LowTargetId = lid
				d.LowTargetType = TYPE_BOT
				if lt == "output" {
					d.LowTargetType = TYPE_BIN
				}
				ch<- d
			}
		}
	}()
	return ch
}

/**
Process recursively checks and executes the bots actions if they are in the ready state to do so. The recursion
runs the same checks against any modified bots.
 */
func Process(id int) {
	b := GetBot(id)
	toProcess := []int{}
	if b.ReadyToSend() {
		//Pass on low value
		if (b.LowTargetType == TYPE_BOT) {
			lb := GetBot(b.LowTargetId)
			lb.Receive(b.Values[0])
			toProcess = append(toProcess, b.LowTargetId)
		} else {
			GetBin(b.LowTargetId).Receive(b.Values[0])
		}
		//Pass on high value
		if (b.HighTargetType == TYPE_BOT) {
			hb := GetBot(b.HighTargetId)
			hb.Receive(b.Values[1])
			toProcess = append(toProcess, b.HighTargetId)
		} else {
			GetBin(b.HighTargetId).Receive(b.Values[1])
		}
		b.Clear()
		for _, tb := range toProcess {
			Process(tb)
		}
	}
}

func GetBin(id int) *bin {
	b, exist := binMap[id]
	if !exist {
		binMap[id] = &bin{Id: id}
		b = binMap[id]
	}
	return b
}

func GetBot(id int) *bot {
	b, exist := botMap[id]
	if !exist {
		botMap[id] = &bot{Id: id}
		b = botMap[id]
	}
	return b
}

/**
Use a goroutine to read from testdata file so that we can
iterate over the lines. Rather than wait for the whole file to be in memory
or embed parsing functions into getInputLines.
*/
func getInputLines() <-chan string {
	ch := make(chan string, 10)
	go func() {
		defer close(ch)
		f, err := os.Open("testdata")
		if err != nil {
			log.Fatal(err)
		}
		r := bufio.NewReader(f)
		for {
			line, err := r.ReadString('\n')
			ch <- strings.TrimSpace(line)
			if err == io.EOF {
				break
			}
		}
	}()
	return ch
}
