package main

import (
	"github.com/rjeczalik/notify"
	"github.com/spf13/viper"
	"log"
	"path"
	"net/smtp"
	"flag"
    "github.com/scorredoira/email"
)

const (
	SUBJECT = "This is the message subject"
	BODY = "This is the message body."
)

func main() {
	var config string
	flag.StringVar(&config, "config", "nfa.toml", "The name of the config file.")
	flag.Parse()
	vconfig := config[:len(config)-len(path.Ext(config))] // strip extension
	viper.SetConfigName(vconfig) 
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Fatal("Fatal error config file: ", err)
	    panic("panic")
	}
	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1)

	// Set up a watchpoint listening on events within current working directory.
	// Dispatch each create and remove events separately to c.
	for {
		// Create, Write, Remove, Rename are events guaranteed to be present in all platforms.
		if err := notify.Watch(viper.GetString("app.directory"), c, notify.Create, notify.Write, notify.Remove, notify.Rename); err != nil {
		    log.Fatal(err)
		}
		defer notify.Stop(c)
		ei := <-c
		log.Println("Got event:", ei)
		// if ei.isCreate() {
			m := email.NewMessage(viper.GetString("mail.subject"), viper.GetString("mail.body"))
			m.From = viper.GetString("mail.from")
			m.To = viper.GetStringSlice("mail.to") // should be []string{"a", "b", ... }
			log.Println("About to attach : " + ei.Path())
			err := m.Attach(ei.Path())
			if err != nil {
			    log.Println(err)
			}
			err = email.Send(viper.GetString("mail.server") + ":" + viper.GetString("mail.port"), smtp.PlainAuth("", viper.GetString("mail.login"), viper.GetString("mail.password"), viper.GetString("mail.server")), m)
			if err != nil {
				log.Println("Can't send mail : ", err)
			}
		// } else {
		//	log.Println("Received a not created file event, ignoring.")
		//}
	}



}