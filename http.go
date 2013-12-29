package pager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/mail"
	"path"

	proto "code.google.com/p/goprotobuf/proto"
	"config"

	"appengine"
	"appengine/datastore"
)

var (
	matchers       = make(map[string]*config.PagerConfig)
	pagers = make(map[string]*config.NotificationSequence)
)

func init() {
	http.HandleFunc("/_ah/mail/", mailHandler)
	http.HandleFunc("/admin/dump_config", dumpConfigHandler)
	http.HandleFunc("/admin/notify", notificationHandler)
	http.HandleFunc("/ack", acknowledgeHandler)
	readNotifications()
	readMatchers()
}

func mailHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	name, tag := parseRecipient(path.Base(r.URL.Path))
	if name == "acknowledge" {
		handleAcknowledge(c, w, r, tag)
		return
	}
	config, ok := matchers[name]
	if !ok {
		c.Errorf("no pager matched name %s", name)
		http.Error(w, "no pager found", 404)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.Errorf("could not read message: %v", err)
		http.Error(w, "", 500)
		return
	}
	msg, err := mail.ReadMessage(bytes.NewReader(body))
	if err != nil {
		c.Errorf("could not parse message: %v", err)
		http.Error(w, "", 500)
		return
	}
	contacts, err := process(msg, config, tag)
	if err != nil {
		c.Errorf("could not process message: %v", err)
		http.Error(w, "", 500)
		return
	}
	pb, err := proto.Marshal(contacts)
	if err != nil {
		c.Errorf("could not marshal pb: %v", err)
		http.Error(w, "", 500)
		return
	}
	escalation := &Escalation{
		Mail:                 body,
		Subject:              []byte(msg.Header.Get("Subject")),
		Body:                 body,
		NotificationSequence: pb,
		LastContact:          -1,
	}
	if err := escalation.save(c); err != nil {
		c.Errorf("could not save escalation: %v", err)
		http.Error(w, "", 500)
		return
	}
	if err := escalation.scheduleNotification(c); err != nil {
		c.Errorf("could not create new task: %v", err)
		http.Error(w, "", 500)
		return
	}
}

func acknowledgeHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	handleAcknowledge(c, w, r, r.FormValue("k"))
}

func handleAcknowledge(c appengine.Context, w http.ResponseWriter, r *http.Request, id string) {
	c.Debugf("handling ACK for %s", id)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		e, err := getEscalation(c, id)
		if err != nil {
			c.Errorf("could not get escalation: %v", err)
			return err
		}
		e.Acknowledged = true
		return e.save(c)
	}, nil)
	if err != nil {
		c.Errorf("could not process ACK: %v", err)
	}
}

func notificationHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		k := r.FormValue("k")
		escalation, err := getEscalation(c, k)
		if err != nil {
			c.Errorf("could not get escalation: %v", err)
			return err
		}
		if escalation.Acknowledged {
			return nil
		}
		seq, err := escalation.GetNotificationSequence(c)
		if err != nil {
			return err
		}
		if escalation.LastContact+1 >= len(seq.Contact) {
			return nil
		}
		contact := *seq.Contact[escalation.LastContact+1]
		if contact.Email != nil {
			if err := sendMail(c, *contact.Email, escalation); err != nil {
				return err
			}
		} else if contact.Sms != nil {
			if err := sendSMS(c, *contact.Sms, escalation); err != nil {
				return err
			}
		} else if contact.Phone != nil {
			if err := sendPhone(c, *contact.Phone, escalation); err != nil {
				return err
			}
		}
		escalation.LastContact++
		return escalation.save(c)
	}, nil)
	if err != nil {
		c.Errorf("error while notifying: %v", err)
		http.Error(w, "", 500)
	}
}

func dumpConfigHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "matchers:\n")
	b, _ := json.MarshalIndent(matchers, "", "\t")
	w.Write(b)
	fmt.Fprintf(w, "\n\npagers:\n")
	b, _ = json.MarshalIndent(pagers, "", "\t")
	w.Write(b)
	proto.CompactText(w, pagers["cbro"])
}
