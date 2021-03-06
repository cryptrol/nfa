package main

import (
	"github.com/rjeczalik/notify"
	"github.com/spf13/viper"
    	"github.com/cryptrol/email"
	"path"
	"log"
	"net/smtp"
	"flag"
)

// Used to search in the available file extensions slice
func SliceContainsString(slice []string, s string) bool {
    for _, i := range slice {
        if i == s {
            return true
        }
    }
    return false
}

func main() {
	var config string
	flag.StringVar(&config, "config", "nfa", "The name of the config file (without the toml extension.")
	flag.Parse()
	viper.SetConfigName(config) 
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/nfa")
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Fatal("Fatal error reading config file.")
	    panic(err)
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
		if ei.Event() == notify.Create {
			if SliceContainsString(viper.GetStringSlice("app.extensions"), path.Ext(ei.Path())[1:]) {
				m := email.NewMessage(viper.GetString("mail.subject"), viper.GetString("mail.body"))
				m.From = viper.GetString("mail.from")
				m.To = viper.GetStringSlice("mail.to") // should be []string{"a", "b", ... }
				log.Println("About to attach : " + ei.Path())
				err := m.Attach(ei.Path())
				if err != nil {
				    log.Println(err)
				}
				addr := viper.GetString("mail.server") + ":" + viper.GetString("mail.port")
				username := viper.GetString("mail.login")
				password := viper.GetString("mail.password")
				host := viper.GetString("mail.server")
				auth := smtp.PlainAuth("", username, password, host)
				if viper.GetBool("mail.useauthlogin") {
					auth = email.LoginAuth(username, password, host)
				}
				skipver := viper.GetBool("mail.nocertverify")
				err = email.Send(addr, auth, m, skipver )
				if err != nil {
					log.Println("Can't send mail : ", err)
				}
			} else {
				log.Println("Ignoring created file with disallowed extension.")
			}
		} else {
			log.Println("Ignoring non create event.")
		}
	}



}
