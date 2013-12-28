# pager

Flexible pager system, written in Go running on AppEngine. E-mails are received via `alias+tag@my-pager.appspotmail.com`. The config for that alias is looked up and a series of pagers are alerted until acknowledeged.

Features:

  * Filtering over tag (e.g. "cpu" in `chris+cpu@`)
  * Filtering by sender
  * Filtering over subject line
  * Filtering over body
  * Substring and regex filters
  * Filtering by time of day
  * Page via e-mail (SMS or phone call coming)
  * Fall-through pager configs (if one channel doesn't acknowledge the alert, the next one will be alerted).

## Configuration

Configuration is currently via text files baked into the binary. The configuration language is the protobuf text format.

There are two sets of configuration: matchers and pagers. Matchers define a set of rules for the incoming message. Pagers are a definition of a series of communication channels (e.g. email, phone, sms) with associated timeouts.

Example:

    # File: matchers/example
    
    rule: <
      # Always send urgent pagers straight to phone.
      # Looks for "URGENT" anywhere in the subject line.
      subject: <
        substring: "URGENT"
      >
      pager: "phone"
    >
    
    rule: <
      # Always send urgent pagers straight to phone.
      # Matches `example+urgent@`
      tag: <
        substring: "urgent"
      >
      pager: "phone"
    >
    
    rule: <
      # During the day, except during lunch, send pagers via e-mail, escalating to phone.
      time: <
        from: 800
        to: 1130
      >
      time: <
        from: 1330
        to: 2000
      >
      # E-mail first, then SMS, then phone.
      pager: "email"
      pager: "sms_then_phone"
    >
    
    rule: <
      # Everything else, only e-mail.
      pager: "email"
    >
    
Example pager configs:

    # File: pagers/phone
    
    contact: <
      phone: "+15551234567"
    >

    # File: pagers/sms_then_phone
    
    contact: <
      # Wait 5 minutes for an ACK, then call the phone.
      timeout: 300
      sms: "+15551234567"
    >
    contact: <
      phone: "+15551234567"
    >

    # File: pagers/email
    
    contact: <
      # Wait 5 minutes for an ACK.
      timeout: 300
      email: "foo@example.com"
    >

## Behaviour/notes

  * A rule set is matched when all of the matchers evaluate to true
  * Rules like subject/sender/time of day may have multiple conditions - these are OR'd
  * If multiple rule sets are defined in a matcher file, the first one to match wins.
