/*
Free slots
1. Read in the calendars to memory*
2. Read in start and end times from prompt*
3. scan meetings.
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"time"
	// "github.com/sqs/goreturns/returns"
)

type Calendar struct {
	Name     string     `json:"name"`
	Meetings []Meetings `json:"meetings"`
}

type Meetings struct {
	SartTime time.Time `json:"starttime"`
	EndTime  time.Time `json:"endtime"`
	Subject  string    `json:"subject"`
}

func failOnError(err error, contex string) {
	if err != nil {
		fmt.Printf("Failed %v with error: %v \n", contex, err)
	}
}

func usage() {
	fmt.Fprintf(
		os.Stderr, `
Zigi FreeSlots is a command line tool for Coordinating calendar slots available for booking. 
		
Usage: 
		./free_slots <sourcefile> <startdate> <enddate>	

sourcefile
		Source file location for calendar entries to be searched. e.g. input.json		

Argumnts are:
	h		Display this usage information
	sourcefile		Source file location for calendar entries to be searched. e.g. input.json
	startdate	Start date to search calendar
	enddate		End date to search calendar

	Zigi FreeSlots

(c) 2021 Zigi all righs reseved.
		`)
}

//DEFINE FLAGS
var (
	Sourcefile, startdate, enddate string
	// h                              = flag.Bool("h", false, "Display usage guide")
	StartDayTime time.Time
	EndDayTime   time.Time
	Cal          []Calendar
	er           error
	mutex        = &sync.Mutex{}
	wg           sync.WaitGroup
)

func flagg(ch chan bool) {

	arg := os.Args

	if len(arg) < 4 {
		fmt.Print("W E L C O M E   T O   Z I g I")
		usage()
		os.Exit(1)
	}

	Sourcefile = arg[1]
	startdate = arg[2]
	enddate = arg[3]

	if len(Sourcefile) > 5 {
		if _, err := os.Stat(Sourcefile); err == nil {
			fmt.Printf("Source file provided : %v \n", Sourcefile)
		} else {
			fmt.Printf("The file %v does not appear to exist. \n", Sourcefile)
			ch <- false
			os.Exit(1)
		}
	} else {
		fmt.Printf("No source file provided \n")
		ch <- false
		os.Exit(1)
	}

	if len(startdate) > 8 {
		//validate date entry
		fmt.Printf("Start Date : %v\n\n", startdate)
		dateLayout := "2006-01-02"

		StartDayTime, er = time.Parse(dateLayout, startdate)
		if er != nil {
			ch <- false
			failOnError(er, "time.Parse startdate")
			os.Exit(1)
		}

	} else {
		ch <- false
		fmt.Printf("Improperly entered startdate : %v\n\n", startdate)
		usage()
		failOnError(er, "Improper startdate")
		os.Exit(1)
	}

	if len(enddate) > 8 {
		//validate date entry
		fmt.Printf("End Date : %v\n\n", enddate)
		dateLayout := "2006-01-02"
		EndDayTime, er = time.Parse(dateLayout, enddate)
		if er != nil {
			ch <- false
			failOnError(er, "time.Parse enddate")
			os.Exit(1)
		}
		// _ = timeToStandard(EndDayTime) //temp

	} else {
		ch <- false
		fmt.Printf("Improperly entered enddate : %v\n\n", enddate)
		usage()
		failOnError(er, "Improper enddate")
		os.Exit(1)
	}

	wrongDaySeq := StartDayTime.After(EndDayTime)
	if wrongDaySeq == true {
		fmt.Printf("Start date must be earlier than end date")
		ch <- false
		os.Exit(1)
	}

	ch <- true
}

var timeMap = make(map[int64]int64)
var timeSlice = make([]int64, 0)
var freeSlotMap = make(map[int64]int64)
var freeSlotSlice = make([]int64, 0)

func getFreeSlots() {

	fmt.Println("Curent Calendar :")
	for _, v := range timeSlice {
		fmt.Printf("StartTime : %v\n", timeToStandard(time.Unix(v, 0)))
		fmt.Printf("EndTime : %v\n\n", timeToStandard(time.Unix(timeMap[v], 0)))
	}

	fmt.Printf("Starting Free Slots..: %v\n", timeSlice)

	var currentStartTime = StartDayTime.Unix()
	currentEndTime := currentStartTime //start checking from the user defined start time

	for x := 1; x <= len(timeSlice); x++ {

		nextStartTime := timeSlice[x-1]
		nextEndTime := timeMap[nextStartTime]

		if nextStartTime > EndDayTime.Unix() {
			//next time slot is not within time frame.
			//end the loop and func
			fmt.Printf("End of time- nextStartTime : %v - EndDayTime : %v \n", nextStartTime, EndDayTime.Unix())
			fmt.Printf("End of time- nextStartTime : %v - EndDayTime : %v \n", time.Unix(nextStartTime, 0), time.Unix(EndDayTime.Unix(), 0))
			return
		}

		if currentEndTime > nextStartTime && currentEndTime > nextEndTime {
			// currentEndTime is currentEndTime //no chage
			//continue
			continue
		}

		if currentEndTime > nextStartTime && currentEndTime < nextEndTime {
			//next end time is nextEndTime
			//
			currentEndTime = nextEndTime
			continue
		}

		if currentEndTime < nextStartTime {
			//free slot is nextStartTime - currentEndTime
			//add to freeSlot slice
			freeSlotSlice = append(freeSlotSlice, currentEndTime)
			freeSlotMap[currentEndTime] = nextStartTime
			currentEndTime = nextEndTime
		}

	}
}

func main() {
	ch := make(chan bool)

	go flagg(ch)
	if <-ch == true {
		c := new(Calendar)
		c.ReadIn(Sourcefile)
	}

	fmt.Printf("Time Map: %v\n\n Time Slice: %v\n\n", timeMap, timeSlice)

	//Remove edundant dates
	for i, v := range timeSlice {
		if v == StartDayTime.Unix() {
			timeSlice = timeSlice[i:]
		}
	}
	getFreeSlots()

	sort.Slice(freeSlotSlice, func(i, j int) bool { return freeSlotSlice[i] < freeSlotSlice[j] })
	fmt.Println("Free Slots:")
	for _, v := range freeSlotSlice {
		fmt.Printf("StartTime : %v \n", timeToStandard(time.Unix(v, 0).UTC()))
		fmt.Printf("EndTime : %v \n\n", timeToStandard(time.Unix(freeSlotMap[v], 0).UTC()))
	}
}

//ReadIn - reads the list of meeting schedule into memory from the given sourcefile.
func (c *Calendar) ReadIn(sourcefile string) {
	fmt.Printf("Attempting to read from: %v \n", sourcefile)

	meetingsData, err := ioutil.ReadFile(sourcefile)
	failOnError(err, "ioutil.ReadFile :"+sourcefile)

	_ = json.Unmarshal([]byte(meetingsData), &Cal)

	mcount := 0
	for _, cal := range Cal {
		wg.Add(1)
		go loadMeetingsMap(cal.Meetings, &wg)
		mcount++
	}
	wg.Wait()

	//append start daytime
	startTimeU := StartDayTime.Unix()
	mutex.Lock()
	timeMap[startTimeU] = StartDayTime.Unix()
	timeSlice = append(timeSlice, startTimeU)
	mutex.Unlock()

	sort.Slice(timeSlice, func(i, j int) bool { return timeSlice[i] < timeSlice[j] })

}

func loadMeetingsMap(m []Meetings, wg *sync.WaitGroup) {

	defer wg.Done()
	for _, t := range m {
		StU := timeToUnix(t.SartTime)
		EtU := timeToUnix(t.EndTime)

		mutex.Lock()
		timeMap[StU] = EtU
		timeSlice = append(timeSlice, StU)
		mutex.Unlock()
	}
}

func timeToStandard(t time.Time) string {
	st := t.Format(time.RFC3339)
	return st
}

func timeToUnix(t time.Time) int64 {
	return t.Unix()
}

//     today := time.Now()
//     tomorrow := today.Add(24 * time.Hour)

//     // Using time.Before() method
//     g1 := today.Before(tomorrow)
//     fmt.Println("today before tomorrow:", g1)

//     // Using time.After() method
//     g2 := tomorrow.After(today)
//     fmt.Println("tomorrow after today:", g2)
