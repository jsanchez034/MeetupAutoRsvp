package mainapp

import (
	"fmt"
	"meetup/meetupautorsvpapp/meetupautorsvp"
	"net/http"
)

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/myevents", myeventshandler)
	http.HandleFunc("/rsvpmyevents", rsvpmyevents)
	http.HandleFunc("/rsvpeventworker", rsvpeventworker)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world! asdf sdf asdf")
}

func myeventshandler(w http.ResponseWriter, r *http.Request) {
	events := meetupautorsvp.GetUpcomingMeetups(r)

	fmt.Fprintf(w, "%+v", events)

}

func rsvpmyevents(w http.ResponseWriter, r *http.Request) {
	events := meetupautorsvp.GetUpcomingMeetups(r)

	str := meetupautorsvp.RSVPMeetupEvents(events, r)
	fmt.Fprint(w, str)
}

func rsvpeventworker(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	meetupautorsvp.PostRSVP(r.PostForm, r)
}
