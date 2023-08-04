package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type InputSettings struct {
	BotName  string `json:"botName,omitempty"`
	BotToken string `json:"botToken,omitempty"`
	ApiUrl   string `json:"apiUrl,omitempty"`
	Pattern  string `json:"pattern"`
}
type BotStatus interface {
	GetStatus()
}

func main() {
	logFile, err := os.OpenFile("log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
	}
	log.SetOutput(logFile)
	inputFile, err := os.OpenFile("settings.json", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
	}
	inputBytes, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Println(err)
	}

	var inputBotsSettings []InputSettings
	err = json.Unmarshal(inputBytes, &inputBotsSettings)
	if err != nil {
		log.Println(err)
	}
	var wg sync.WaitGroup

	for i := 0; i < len(inputBotsSettings); i++ {
		wg.Add(1)
		go func(i int) {
			fmt.Println(fmt.Sprintf("Bot %s, works", inputBotsSettings[i].BotName))
			gd, err := discordgo.New(inputBotsSettings[i].BotToken)
			if err != nil {
				log.Println(err)
			}

			err = gd.Open()
			if err != nil {
				log.Println(err)
			}

			for {
				data := getDayzStatusServer(inputBotsSettings[i].ApiUrl, inputBotsSettings[i].Pattern)
				err := gd.UpdateGameStatus(0, data)
				if err != nil {
					log.Println(err)
				}
				time.Sleep(10 * time.Second)
				debug.FreeOSMemory()
			}
		}(i)

	}
	wg.Wait()
}

func getDayzStatusServer(api string, pattern string) string {

	resp, err := http.Get(api)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)
	}
	var result string

	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {

		playersStr := s.Find("td").Eq(1).Text()
		queueStr := s.Find("td").Eq(5).Text()

		// Extract the number of players and queue from the strings
		playersParts := strings.Split(playersStr, "/")
		if len(playersParts) != 2 {
			log.Printf("Invalid players string: %s", playersStr)
			return
		}

		serverTimeStr := doc.Find("tbody tr td").Eq(3).Text()
		result = strings.ReplaceAll(pattern, "%time", serverTimeStr)
		result = strings.ReplaceAll(result, "%maxPlayers", playersParts[1])
		result = strings.ReplaceAll(result, "%numPlayers", playersParts[0])
		result = strings.ReplaceAll(result, "%queue", queueStr)

	})

	return result
}
