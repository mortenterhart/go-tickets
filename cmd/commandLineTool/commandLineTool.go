package main

import (
	"flag"
	"github.com/mortenterhart/trivial-tickets/IO"
	"github.com/mortenterhart/trivial-tickets/communicationToServer"
	"github.com/mortenterhart/trivial-tickets/structs"
	"github.com/mortenterhart/trivial-tickets/util/cliUtils"
	"log"
	"math"
)

func main() {

	conf, fetch, submit, mail := getConfig()
	communicationToServer.SetServerConfig(conf)
	switch {
	case (!submit && !fetch):
		commandLoop()
	case (!submit && fetch):
		mails, err := communicationToServer.FetchEmails()
		if err != nil {
			log.Fatal(err)
		}
		for _, mail := range mails {
			IO.PrintEmail(mail)
		}
	case (submit && !fetch):
		communicationToServer.SubmitEmail(mail)
	default:
		log.Fatal(structs.NoValidOption)
	}

}

// Here's how it should behave:
// On startup, show the available commands (fetch, submit and exit)
// On fetch just print out all available emails in the command line.
// ON submit ask in order for email, ticketID, subject and message.
// After both fetch and submit, return to startup message.
// On exit, close the application.

// When starting the Application it should be possible to specify some parameters.
// Based on these parameters the IP address of the server and its port should be selected.
// The Parameters should also offer a way to directly invoke the send / fetch mail commands.

func getConfig() (conf structs.CLIConfig, fetch bool, submit bool, mail structs.Mail) {
	IPAddr := flag.String("ip", "127.0.0.0", "IP address of the server")
	port := flag.Uint("port", 443, "Port the server listens to")
	f := flag.Bool("f", false, "fetch (fetch): If set, the application will fetch all messages from the server.")
	s := flag.Bool("s", false, "Use to submit a message to the server. Requires -email, -subject, -message. The use of -tID is optional.")
	email := flag.String("email", "", "The eamil address of the sender")
	ticketID := flag.String("tID", "", "ID of the related Ticket. If left empty, a new ticket is created")
	subject := flag.String("subject", "", "The subject of the message")
	message := flag.String("message", "", "The body of the message.")

	flag.Parse()

	if *port > math.MaxUint16 {
		log.Fatal("Port is not a valid port number.")
	}

	conf = structs.CLIConfig{
		IPAddr: *IPAddr,
		Port:   uint16(*port),
	}

	fetch = *f
	submit = *s
	mail = cliUtils.CreateMail(*email, *ticketID, *subject, *message)
	return
}

func commandLoop() {
	ok := true
	for ok {
		com, err := IO.NextCommand()
		if err != nil {
			log.Fatal(err)
		}
		switch com {
		case structs.FETCH:
			mails, err := communicationToServer.FetchEmails()
			if err != nil {
				log.Fatal(err)
			}
			for _, mail := range mails {
				IO.PrintEmail(mail)
			}
		case structs.SUBMIT:
			mail, err := IO.GetEmail()
			if err != nil {
				log.Fatal(err)
			}
			communicationToServer.SubmitEmail(mail)
		case structs.EXIT:
			ok = false
		default:
			log.Fatal(com, err)
		}
	}
}