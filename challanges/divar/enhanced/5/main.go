package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	freeFloorType    = "free"
	specialFloorType = "special"

	multipleReserveCommand = "reserve_multiple_desks"
	reserveCommand         = "reserve_desk"
	requestCommand         = "request_desk"
)

var (
	reader *bufio.Reader

	featuresPrice map[int]int

	requestDeskPrice    = "%s got desk %d-%d for %d"
	reserveDesk         = "%s reserved desk %d-%d for %d"
	reserveMultipleDesk = "%s reserved desk %s for %d"

	noDesk         = "no desk available"
	notEnoughDesks = "not enough desks available"
)

type Command struct {
	username  string
	timestamp int
	fromTime  int
	duration  int
}

func (com1 *Command) crossBoundaries(com2 *Command) bool {
	start1 := com1.timestamp
	end1 := com1.timestamp + com1.duration
	start2 := com2.timestamp
	end2 := com2.timestamp + com2.duration

	startCond := start2 > start1 && start2 < end1
	endCond := end2 > start1 && end2 < end1
	containsCond := start2 <= start1 && end2 >= end1

	if startCond || endCond || containsCond {
		return false
	}

	return true

}

type Reserve struct {
	command  Command
	features []int
}

type Request struct {
	command   Command
	isSpecial bool
}

type Desk struct {
	number   int
	features []int
	commands []Command
}

func (d *Desk) canReserveMultiple(reserve *Reserve) bool {
	for index := 0; index < len(d.commands); index++ {
		if cross := d.commands[index].crossBoundaries(&reserve.command); !cross {
			return false
		}
	}

	return true
}

func (d *Desk) canReserve(reserve *Reserve) bool {
	if !areEqual(d.features, reserve.features) {
		return false
	}

	for index := 0; index < len(d.commands); index++ {
		if cross := d.commands[index].crossBoundaries(&reserve.command); !cross {
			return false
		}
	}

	return true
}

func (d *Desk) canRequest(request *Request) bool {
	for index := 0; index < len(d.commands); index++ {
		if cross := d.commands[index].crossBoundaries(&request.command); !cross {
			return false
		}
	}

	return true
}

type Floor struct {
	number int
	desks  []Desk
	price  int
}

func (floor *Floor) reserveMultiple(reserve *Reserve, count int) []*Desk {
	if floor.price == 0 {
		return nil
	}

	desks := make([]*Desk, 0, count)

	for index := 0; index < len(floor.desks); index++ {
		if floor.desks[index].canReserveMultiple(reserve) {
			desks = append(desks, &floor.desks[index])

			if len(desks) == count {
				break
			}
		}
	}

	if len(desks) == count {
		for index := 0; index < count; index++ {
			desks[index].commands = append(desks[index].commands, reserve.command)
		}

		return desks
	}

	return nil
}

func (floor *Floor) reserve(reserve *Reserve) *Desk {
	if floor.price == 0 {
		return nil
	}

	var desk *Desk

	for index := 0; index < len(floor.desks); index++ {
		if floor.desks[index].canReserve(reserve) {
			desk = &floor.desks[index]
			break
		}
	}

	if desk != nil {
		desk.commands = append(desk.commands, reserve.command)
	}

	return desk
}

func (floor *Floor) request(req *Request) *Desk {
	cond1 := !req.isSpecial && floor.price != 0
	cond2 := req.isSpecial && floor.price == 0

	if cond1 || cond2 {
		return nil
	}

	var desk *Desk

	for index := 0; index < len(floor.desks); index++ {
		if floor.desks[index].canRequest(req) {
			desk = &floor.desks[index]
			break
		}
	}

	if desk != nil {
		desk.commands = append(desk.commands, req.command)
	}

	return desk
}

type Floors []Floor

func (floors Floors) reserveMultiple(reserve *Reserve, count int) {
	for floorIndex := 0; floorIndex < len(floors); floorIndex++ {
		floor := floors[floorIndex]

		if desks := floor.reserveMultiple(reserve, count); desks != nil {
			var text string

			featurePrice := 0
			for index1 := 0; index1 < count; index1++ {
				for index2 := 0; index2 < len(desks[index1].features); index2++ {
					featurePrice += featuresPrice[desks[index1].features[index2]]
				}
			}

			var sb strings.Builder
			for index := 0; index < count; index++ {
				txt := fmt.Sprintf("%d-%d ", floor.number, desks[index].number)
				sb.WriteString(txt)

			}

			text = fmt.Sprintf(
				reserveMultipleDesk,
				reserve.command.username,
				strings.Trim(sb.String(), " "),
				count*floor.price+reserve.command.duration*featurePrice,
			)

			fmt.Println(text)
			return
		}
	}

	fmt.Println(notEnoughDesks)
}

func (floors Floors) reserve(reserve *Reserve) {
	for floorIndex := 0; floorIndex < len(floors); floorIndex++ {
		floor := floors[floorIndex]

		if desk := floor.reserve(reserve); desk != nil {
			var text string

			featurePrice := 0
			for index := 0; index < len(desk.features); index++ {
				featurePrice += featuresPrice[desk.features[index]]
			}

			text = fmt.Sprintf(
				reserveDesk,
				reserve.command.username,
				floor.number,
				desk.number,
				floor.price+reserve.command.duration*featurePrice,
			)

			fmt.Println(text)
			return
		}
	}

	fmt.Println(noDesk)
}

func (floors Floors) request(req *Request) {
	for index := 0; index < len(floors); index++ {
		floor := floors[index]

		if desk := floor.request(req); desk != nil {
			var text string

			featurePrice := 0
			for index := 0; index < len(desk.features); index++ {
				featurePrice += featuresPrice[desk.features[index]]
			}

			text = fmt.Sprintf(
				requestDeskPrice,
				req.command.username,
				floor.number,
				desk.number,
				floor.price+req.command.duration*featurePrice,
			)

			fmt.Println(text)
			return
		}
	}

	fmt.Println(noDesk)
}

func main() {
	reader = bufio.NewReader(os.Stdin)
	featuresPrice = map[int]int{}

	processInput()
}

func readLine() string {
	line, _ := reader.ReadString('\n')
	return strings.Trim(line, "\n")
}

func processInput() {
	line := readLine()
	featuresNum, _ := strconv.Atoi(line)

	line = readLine()
	values := strings.Split(line, " ")
	for index := 0; index < featuresNum; index++ {
		price, _ := strconv.Atoi(values[index])
		featuresPrice[index] = price
	}

	line = readLine()
	values = strings.Split(line, " ")
	floorsNum, _ := strconv.Atoi(values[0])
	specialPrice, _ := strconv.Atoi(values[1])

	floors := make(Floors, 0, floorsNum)

	for index := 0; index < floorsNum; index++ {
		floor := Floor{number: index + 1, desks: make([]Desk, 0)}

		line := readLine()
		values := strings.Split(line, " ")
		desksNum, _ := strconv.Atoi(values[0])
		floorType := values[1]

		if floorType == specialFloorType {
			floor.price = specialPrice
		}

		line = readLine()
		values = strings.Split(line, " ")

		for index := 0; index < desksNum; index++ {
			floor.desks = append(
				floor.desks,
				Desk{
					number:   index + 1,
					commands: make([]Command, 0),
					features: stringToFeatures(values[index]),
				},
			)
		}

		floors = append(floors, floor)
	}

	for {
		line := readLine()

		if line == "end" {
			break
		}

		values := strings.Split(line, " ")
		timestamp, _ := strconv.Atoi(values[0])
		username := values[2]

		command := Command{
			username:  username,
			timestamp: timestamp,
			fromTime:  timestamp,
		}

		if values[1] == requestCommand {
			duration, _ := strconv.Atoi(values[4])
			command.duration = duration

			request := Request{
				command:   command,
				isSpecial: values[3] == specialFloorType,
			}

			floors.request(&request)
		} else if values[1] == reserveCommand {
			fromTime, _ := strconv.Atoi(values[3])
			duration, _ := strconv.Atoi(values[4])
			command.fromTime = fromTime
			command.duration = duration
			reserve := Reserve{
				command:  command,
				features: stringToFeatures(values[5]),
			}

			floors.reserve(&reserve)
		} else {
			count, _ := strconv.Atoi(values[3])
			fromTime, _ := strconv.Atoi(values[4])
			duration, _ := strconv.Atoi(values[5])
			command.fromTime = fromTime
			command.duration = duration
			reserve := Reserve{
				command: command,
			}

			floors.reserveMultiple(&reserve, count)
		}
	}
}

func areEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func stringToFeatures(input string) []int {
	result := make([]int, 0)

	for index := 0; index < len(input); index++ {
		if string(input[index]) == "1" {
			result = append(result, index)
		}
	}

	return result
}
