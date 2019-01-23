// Mail message construction using templating
package mail_events

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"

	"github.com/mortenterhart/trivial-tickets/globals"
	"github.com/mortenterhart/trivial-tickets/logger"
	"github.com/mortenterhart/trivial-tickets/structs"
)

/*
 * Ticketsystem Trivial Tickets
 *
 * Matriculation numbers: 3040018, 6694964, 3478222
 * Lecture:               Programmieren II, INF16B
 * Lecturer:              Herr Prof. Dr. Helmut Neemann
 * Institute:             Duale Hochschule Baden-Württemberg Mosbach
 *
 * ---------------
 *
 * Package mail_events
 * Mail message construction using templating
 */

type Event int

type templateMap map[string]string

const (
	NewTicket Event = iota
	NewAnswer
	UpdatedTicket
	AssignedTicket
	UnassignedTicket
)

func (event Event) String() string {
	switch event {
	case NewTicket:
		return "new ticket"

	case NewAnswer:
		return "new answer"

	case UpdatedTicket:
		return "updated ticket"

	case AssignedTicket:
		return "assigned ticket"

	case UnassignedTicket:
		return "unassigned ticket"
	}

	return "undefined"
}

// NewMailBody creates a message to be sent inside a mail body.
// Depending of the mail event (e.g. ticket or answer creation)
// different messages are written to the body and populated with
// information from a given ticket.
func NewMailBody(event Event, ticket structs.Ticket) string {
	mailTemplate := template.New("mail_body")

	// The string was originally built using strings.Builder, however
	// this type was firstly introduced in Go 1.10
	var mailBuilder bytes.Buffer
	mailBuilder.WriteString("Sehr geehrter Kunde, sehr geehrte Kundin,\n\n")

	displayLatestAnswer := false

	var eventMessage string
	switch event {
	case NewTicket:
		eventMessage = "Ihr Ticket '{{.ticketId}}' ist erfolgreich erstellt worden.\n" +
			fmt.Sprintf("Wenn Sie eine neuen Kommentar zu diesem Ticket schreiben wollen,\n"+
				"nutzen Sie bitte den folgenden Link: mailto:support@trivial-tickets.com?subject=%s",
				url.PathEscape(fmt.Sprintf(`[Ticket "%s"] %s`, ticket.Id, ticket.Subject)))

	case NewAnswer:
		displayLatestAnswer = true
		eventMessage = "der Benutzer '{{.newAnswerUser}}' hat einen neuen Kommentar geschrieben:\n"

	case UpdatedTicket:
		eventMessage = "Ihr Ticket '{{.ticketId}}' wurde mit folgenden Informationen aktualisiert:\n"

	case AssignedTicket:
		eventMessage = "unser Mitarbeiter '{{.assignedUserName}}' bearbeitet nun Ihr Ticket:\n"

	case UnassignedTicket:
		eventMessage = "der Bearbeiter '{{.assignedUserName}}' hat das Ticket wieder freigegeben:\n"

	}

	mailBuilder.WriteString(eventMessage)

	templateString := `
-----------------------------
Kunde:      {{.customer}}
Schlüssel:  {{.ticketId}}
URL:        {{.url}}
Bearbeiter: {{.assignedUserMail}}
Status:     {{.status}}

Betreff: {{.subject}}

{{.message}}

-----------------------------

Mit freundlichen Grüßen
Ihr trivial-tickets Team

Diese Meldung wurde automatisch durch trivial-tickets.com generiert.
Bitte antworten Sie nicht auf diese E-Mail.`

	mailBuilder.WriteString(templateString)

	parsedTemplate, parseErr := mailTemplate.Parse(mailBuilder.String())
	if parseErr != nil {
		logger.Error("internal error: could not build mail message from template:", parseErr)
		return ""
	}

	mailBuilder.Reset()

	message, newAnswerUser := getMessage(ticket.Entries, displayLatestAnswer)

	assignedUserMail, assignedUserName := getAssignedUser(ticket.User)

	executeErr := parsedTemplate.Execute(&mailBuilder, templateMap{
		"customer":         ticket.Customer,
		"ticketId":         ticket.Id,
		"url":              fmt.Sprintf("https://localhost:%d/ticket?id=%s", globals.ServerConfig.Port, ticket.Id),
		"assignedUserName": assignedUserName,
		"assignedUserMail": assignedUserMail,
		"status":           ticket.Status.String(),
		"subject":          ticket.Subject,
		"message":          message,
		"newAnswerUser":    newAnswerUser,
	})

	if executeErr != nil {
		logger.Error("internal error: could not fill mail template with ticket information:", executeErr)
		return ""
	}

	return mailBuilder.String()
}

func getAssignedUser(user structs.User) (string, string) {
	if user == (structs.User{}) {
		return "kein Bearbeiter zugewiesen", "<nicht zugewiesen>"
	}

	return fmt.Sprintf("%s (%s)", user.Name, user.Mail), user.Name
}

func getMessage(ticketEntries []structs.Entry, displayLatestMessage bool) (string, string) {
	if len(ticketEntries) == 0 {
		return "kein Eintrag vorhanden", ""
	}

	if displayLatestMessage {
		latestEntry := ticketEntries[len(ticketEntries)-1]
		return latestEntry.Text, latestEntry.User
	}

	return ticketEntries[0].Text, ""
}
