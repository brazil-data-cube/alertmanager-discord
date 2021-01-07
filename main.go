package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

// Discord color values
const (
	ColorRed   = 10038562
	ColorGreen = 3066993
	ColorGrey  = 9807270
)

type alertManAlert struct {
	Annotations struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"annotations"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
	StartsAt     string            `json:"startsAt"`
	Status       string            `json:"status"`
}

type alertManOut struct {
	Alerts            []alertManAlert `json:"alerts"`
	CommonAnnotations struct {
		Summary string `json:"summary"`
	} `json:"commonAnnotations"`
	CommonLabels struct {
		Alertname string `json:"alertname"`
	} `json:"commonLabels"`
	ExternalURL string `json:"externalURL"`
	GroupKey    string `json:"groupKey"`
	GroupLabels struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Version  string `json:"version"`
}

type discordOut struct {
	Content string         `json:"content"`
	Embeds  []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Fields      []discordEmbedField `json:"fields"`
}

type discordEmbedField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

const defaultListenAddress = "127.0.0.1:9094"

func main() {

	// group marked to receive the alerts
	alertGroup := os.Getenv("ALERT_GROUP_ID")

	envWhURL := os.Getenv("DISCORD_WEBHOOK")
	whURL := flag.String("webhook.url", envWhURL, "Discord WebHook URL.")

	envListenAddress := os.Getenv("LISTEN_ADDRESS")
	listenAddress := flag.String("listen.address", envListenAddress, "Address:Port to listen on.")

	flag.Parse()

	if *whURL == "" {
		log.Fatalf("Environment variable 'DISCORD_WEBHOOK' or CLI parameter 'webhook.url' not found.")
	}

	if *listenAddress == "" {
		*listenAddress = defaultListenAddress
	}

	_, err := url.Parse(*whURL)
	if err != nil {
		log.Fatalf("The Discord WebHook URL doesn't seem to be a valid URL.")
	}

	re := regexp.MustCompile(`https://discord(?:app)?.com/api/webhooks/[0-9]{18}/[a-zA-Z0-9_-]+`)
	if ok := re.Match([]byte(*whURL)); !ok {
		log.Printf("The Discord WebHook URL doesn't seem to be valid.")
	}

	log.Printf("Listening on: %s", *listenAddress)
	http.ListenAndServe(*listenAddress, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		amo := alertManOut{}
		err = json.Unmarshal(b, &amo)
		if err != nil {
			panic(err)
		}

		groupedAlerts := make(map[string][]alertManAlert)

		for _, alert := range amo.Alerts {
			groupedAlerts[alert.Status] = append(groupedAlerts[alert.Status], alert)
		}

		for status, alerts := range groupedAlerts {
			DO := discordOut{}

			embedTitle := ""
			if status == "firing" {
				embedTitle = fmt.Sprintf("New %s alert", amo.CommonLabels.Alertname)
			} else if status == "resolved" {
				embedTitle = fmt.Sprintf("%s resolved", amo.CommonLabels.Alertname)
			}

			RichEmbed := discordEmbed{
				Title:       embedTitle,
				Description: amo.CommonAnnotations.Summary,
				Color:       ColorGrey,
				Fields:      []discordEmbedField{},
			}

			if status == "firing" {
				RichEmbed.Color = ColorRed

				DO.Content = fmt.Sprintf("%s There is a new prometheus alert that needs attention :fire:", alertGroup)
			} else if status == "resolved" {
				RichEmbed.Color = ColorGreen

				DO.Content = fmt.Sprintf("%s I bring you good news :partying_face: A problem alert generated on prometheus has been solved.", alertGroup)
			}

			for _, alert := range alerts {
				enbededName := ""
				enbededValue := ""

				if status == "firing" {
					enbededName = fmt.Sprintf("[Problem] %s on %s", alert.Labels["alertname"], alert.Labels["instance_name"])

					enbededValue = fmt.Sprintf("I am sending this message to inform you that Prometheus target %s is experiencing problems related to %s. The criticality of this event is classified as %s. Make sure everything is correct.", alert.Labels["instance_name"], alert.Labels["alertname"], alert.Labels["severity"])
				} else if status == "resolved" {
					enbededName = fmt.Sprintf("[Solved] %s on %s has been solved", alert.Labels["alertname"], alert.Labels["instance_name"])

					enbededValue = fmt.Sprintf("I am sending this message to inform you that the problem reported on Prometheus target %s has just been solved, and everything is now under control.", alert.Labels["instance_name"])
				}

				RichEmbed.Fields = append(RichEmbed.Fields, discordEmbedField{
					Name:  enbededName,
					Value: enbededValue,
				})
			}

			DO.Embeds = []discordEmbed{RichEmbed}

			DOD, _ := json.Marshal(DO)
			http.Post(*whURL, "application/json", bytes.NewReader(DOD))
		}
	}))
}
