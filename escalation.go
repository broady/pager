package pager

import (
	"net/url"
	"time"

	"config"
	proto "github.com/golang/protobuf/proto"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
)

type Escalation struct {
	key                  *datastore.Key
	Mail                 []byte
	Subject              []byte
	Body                 []byte
	NotificationSequence []byte
	LastContact          int
	Acknowledged         bool
}

func getEscalation(c appengine.Context, id string) (*Escalation, error) {
	k, err := datastore.DecodeKey(id)
	if err != nil {
		return nil, err
	}
	esc := new(Escalation)
	if err = datastore.Get(c, k, esc); err != nil {
		return nil, err
	}
	esc.key = k
	return esc, nil
}

func (e *Escalation) scheduleNotification(c appengine.Context) (err error) {
	seq, err := e.GetNotificationSequence(c)
	if err != nil {
		return err
	}
	t := taskqueue.NewPOSTTask("/admin/notify", url.Values{
		"k": []string{e.key.Encode()},
	})
	if e.LastContact > -1 {
		t.Delay = time.Second * time.Duration(seq.Contact[e.LastContact].GetTimeout())
	}
	if _, err := taskqueue.Add(c, t, ""); err != nil {
		return err
	}
	return nil
}

func (e *Escalation) save(c appengine.Context) (err error) {
	if e.key == nil {
		e.key = datastore.NewIncompleteKey(c, "Escalation", nil)
	}
	e.key, err = datastore.Put(c, e.key, e)
	return err
}

func (e *Escalation) GetNotificationSequence(c appengine.Context) (ns *config.NotificationSequence, err error) {
	seq := new(config.NotificationSequence)
	if err := proto.Unmarshal(e.NotificationSequence, seq); err != nil {
		return nil, err
	}
	return seq, nil
}
