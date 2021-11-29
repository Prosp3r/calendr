package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"time"
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

type Result struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
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
	StartDayTime                   time.Time
	EndDayTime                     time.Time
	Cal                            []Calendar
	er                             error
	mutex                          = &sync.Mutex{}
	wg                             sync.WaitGroup
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
		fmt.Printf("Start Date : %v\n", startdate)
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
		fmt.Printf("End Date : %v\n", enddate)
		dateLayout := "2006-01-02"
		EndDayTime, er = time.Parse(dateLayout, enddate)
		if er != nil {
			ch <- false
			failOnError(er, "time.Parse enddate")
			os.Exit(1)
		}

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

	var currentStartTime = StartDayTime.Unix()
	currentEndTime := currentStartTime

	for x := 1; x <= len(timeSlice); x++ {

		nextStartTime := timeSlice[x-1]
		nextEndTime := timeMap[nextStartTime]

		if nextEndTime > EndDayTime.Unix() || nextStartTime > EndDayTime.Unix() {
			break
		}

		if currentEndTime > nextStartTime && currentEndTime > nextEndTime {
			continue
		}

		if currentEndTime > nextStartTime && currentEndTime < nextEndTime {
			currentEndTime = nextEndTime
			continue
		}

		if currentEndTime < nextStartTime {
			freeSlotSlice = append(freeSlotSlice, currentEndTime)
			freeSlotMap[currentEndTime] = nextStartTime
			currentEndTime = nextEndTime
		}

		if x == len(timeSlice) && currentEndTime < EndDayTime.Unix() {
			freeSlotSlice = append(freeSlotSlice, currentEndTime)
			freeSlotMap[currentEndTime] = EndDayTime.Unix()
			currentEndTime = EndDayTime.Unix()
		}

	}
}

func main() {
	ch := make(chan bool)

	go flagg(ch)
	if <-ch == true {
		// fmt.Printf("Start Date Converted: %v\n\n", StartDayTime.UTC())
		// fmt.Printf("Start Date ConvertedUnix: %v\n\n", StartDayTime.Unix())
		// fmt.Printf("Start Date ConvertedfromUnix: %v\n\n", time.Unix(StartDayTime.UTC().Unix(), 0).UTC())

		// fmt.Printf("Start Date ConvertedfromUnix: %v\n\n", timeToStandard(time.Unix(StartDayTime.Unix(), 0)))
		// fmt.Printf("Standard Start Date Converted: %v\n\n", timeToStandard(StartDayTime))
		// fmt.Printf("Standard Start Date Converted: %v\n\n", timeToStandard(StartDayTime.UTC()))
		c := new(Calendar)
		c.ReadIn(Sourcefile)
	}

	//Remove redundant dates
	for i, v := range timeSlice {
		if v == StartDayTime.Unix() {
			timeSlice = timeSlice[i:]
		}
	}
	getFreeSlots()

	sort.Slice(freeSlotSlice, func(i, j int) bool { return freeSlotSlice[i] < freeSlotSlice[j] })

	var results []Result
	for _, v := range freeSlotSlice {
		StartTime := timeToStandard(time.Unix(v, 0).UTC())
		EndTime := timeToStandard(time.Unix(freeSlotMap[v], 0).UTC())
		result := Result{
			StartTime: StartTime,
			EndTime:   EndTime,
		}
		results = append(results, result)
	}
	r, err := json.Marshal(results)
	failOnError(err, "Failed marshalling result to json")
	fmt.Printf("\n\nFree Slots: \n%v\n", string(r))
}

//ReadIn - reads the list of meeting schedule into memory from the given sourcefile.
func (c *Calendar) ReadIn(sourcefile string) {

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
	timeMap[startTimeU] = EndDayTime.Unix()
	timeSlice = append(timeSlice, startTimeU)
	mutex.Unlock()

	sort.Slice(timeSlice, func(i, j int) bool { return timeSlice[i] < timeSlice[j] })
}

func loadMeetingsMap(m []Meetings, wg *sync.WaitGroup) {

	defer wg.Done()
	for _, t := range m {
		StU := t.SartTime.Unix()
		EtU := t.EndTime.Unix()

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
