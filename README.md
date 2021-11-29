# Calendr

### A simple commandline calendar time slot search

### Usage

```$ ./free_slots```

Will dispaly usage instructions


Command style:
```./free_slots <sourcefile> <startdate> <enddate>```

e.g.
```$ ./free_slots input.json 2021-03-08 2021-03-12```


This will display available/free slots of time in the calendar

```
[
 {
  "startTime": "2021-03-08T00:00:00Z",
  "endTime": "2021-03-10T08:15:31Z"
 },
 {
  "startTime": "2021-03-10T10:15:13Z",
  "endTime": "2021-03-10T11:55:31Z"
 },
 {
  "startTime": "2021-03-10T12:15:19Z",
  "endTime": "2021-03-11T10:15:45Z"
 },
 {
  "startTime": "2021-03-11T10:55:28Z",
  "endTime": "2021-03-11T11:15:51Z"
 },
 {
  "startTime": "2021-03-11T12:55:14Z",
  "endTime": "2021-03-12T00:00:00Z"
 }
]
```


SIDE NOTE: 
- Time Zone is default to UTC for consistency.