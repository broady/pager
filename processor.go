package pager

import (
	"fmt"
	"io/ioutil"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"config"
)

func process(page *mail.Message, c *config.PagerConfig, tag string) (*config.NotificationSequence, error) {
	for _, rule := range c.Rule {
		if len(rule.Sender) > 0 && !matchString(rule.Sender, page.Header.Get("From")) {
			continue
		}
		if len(rule.Time) > 0 && !matchTimeRange(rule.Time, time.Now()) {
			continue
		}
		if len(rule.Subject) > 0 && !matchString(rule.Subject, page.Header.Get("Subject")) {
			continue
		}
		if len(rule.Body) > 0 {
			body, err := ioutil.ReadAll(page.Body)
			if err != nil {
				return nil, err
			}
			if !matchString(rule.Body, string(body)) {
				continue
			}
		}
		if len(rule.Tag) > 0 && !matchString(rule.Tag, tag) {
			continue
		}
		return getNotificationSequence(rule)
	}
	// todo: error.
	return nil, nil
}

func matchTimeRange(tr []*config.TimeRange, t time.Time) bool {
	timeInt := uint32(t.Hour()*100 + t.Minute())
	for _, r := range tr {
		if timeInt > *r.From && timeInt <= *r.To {
			return true
		}
	}
	return false
}

func matchString(sm []*config.StringMatch, s string) bool {
	for _, m := range sm {
		if m.Regex != nil && regexp.MustCompile(*m.Substring).MatchString(s) {
			return true
		}
		if m.Substring != nil && strings.Contains(s, *m.Substring) {
			return true
		}
	}
	return false
}

func getNotificationSequence(rule *config.Matcher) (*config.NotificationSequence, error) {
	var cs []*config.Contact
	for _, id := range rule.Pager {
		if ns, ok := pagers[id]; !ok {
			return nil, fmt.Errorf("contacts not found: %s", id)
		} else {
			cs = append(cs, ns.Contact...)
		}
	}
	return &config.NotificationSequence{Contact: cs}, nil
}

func parseRecipient(email string) (name, tag string) {
	first := strings.Split(email, "@")[0]
	parts := strings.Split(first, "+")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
