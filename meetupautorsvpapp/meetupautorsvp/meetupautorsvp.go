package meetupautorsvp

import (
	"appengine"
	"appengine/taskqueue"
	"appengine/urlfetch"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apikey string = ""
const memberid string = ""

type MeetupMeta struct {
	Lon         string
	Count       int64
	Signed_url  string
	Link        string
	Next        string
	Total_count int64
	Url         string
	Id          string
	Title       string
	Update      int64
	Description string
	Method      string
	Lat         string
}

type MeetupEventResults struct {
	Results []MeetupEvent
	Meta    MeetupMeta
}

type MeetupEvent struct {
	Rsvp_limit       int64
	Status           string
	Visibility       string
	Maybe_rsvp_count int64
	Venue            interface{}
	Rsvp_rules       RsvpRules
	Fee              interface{}
	Id               string
	Utc_offset       int64
	Duration         int64
	Time             int64
	Waitlist_count   int64
	Updated          int64
	Yes_rsvp_count   int64
	Created          int64
	Event_url        string
	Name             string
	Group            interface{}
}

type RsvpRules struct {
	Close_time    int64
	Closed        int
	Open_time     int64
	Refund_policy interface{}
}

func LogError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetMyMeetupEvents(r *http.Request) (MeetupEventResults, error) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.Get("https://api.meetup.com/2/events?key=" + apikey + "&rsvp=none&member_id=" + memberid + "&fields=rsvp_rules&page=20")
	LogError(err)

	var buf []byte
	var result MeetupEventResults

	if err != nil {
		return result, err
	}

	buf, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	LogError(err)

	json.Unmarshal(buf, &result)

	return result, nil
}

func ProcessEvent(event MeetupEvent, r *http.Request) *taskqueue.Task {
	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	now := time.Now().Unix() * 1000
	opentime := event.Rsvp_rules.Open_time

	var task *taskqueue.Task = nil

	data := url.Values{"event_id": {event.Id}, "agree_to_refund": {"false"}, "rsvp": {"yes"}}

	if opentime <= now {
		PostRSVP(data, r)
		t := time.Unix(event.Time/1000, 0)
		log.Printf("RSVPing for %v on %s\n", event.Name, t.Format(layout))
	} else if opentime > now {
		task = taskqueue.NewPOSTTask("/rsvpeventworker", data)
		task.ETA = time.Unix(opentime/1000, 0)
	}

	return task

}

func PostRSVP(data url.Values, r *http.Request) {
	const layout = "Jan 2, 2006 at 3:04pm (MST)"

	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	resp, err := client.Post("https://api.meetup.com/2/rsvp?key="+apikey,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))

	var buf []byte
	var result interface{}

	if err != nil {
		LogError(err)
		return
	}

	buf, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	json.Unmarshal(buf, &result)

	log.Printf("RSVP Post Result: %+v\n", result)
}

func GetUpcomingMeetups(r *http.Request) []MeetupEvent {
	result, _ := GetMyMeetupEvents(r)

	return result.Results
}

func RSVPMeetupEvents(events []MeetupEvent, r *http.Request) string {
	c := appengine.NewContext(r)
	queuetasks := make([]*taskqueue.Task, 0)

	taskqueue.Purge(c, "futurersvp")

	for _, event := range events {
		t := ProcessEvent(event, r)
		if t != nil {
			queuetasks = append(queuetasks, t)
		}
	}

	taskqueue.AddMulti(c, queuetasks, "futurersvp")

	str := fmt.Sprintf("Total Number of Meetups processed: %d\n", len(events))
	str = str + fmt.Sprintf("Total Number of future Meetups queued: %d\n", len(queuetasks))
	log.Print(str)

	return str
}
