package main

import (
	"net/http"
	"log"
	"ynoserver/server"
	"strconv"
	"io/ioutil"
	"flag"
	"encoding/json"
	"gopkg.in/natefinch/lumberjack.v2"
	"strings"
	"fmt"
)

func writeLog(ip string, payload string, errorcode int) {
	log.Printf("%v \"%v\" %v\n", ip, payload, errorcode)
}

func contains(s []string, num int) bool {
	for _, v := range s {
		if v == strconv.Itoa(num) {
			return true
		}
	}

	return false
}

func main() {
	config_file := flag.String("config", "config.yml", "Path to the configuration file")
	flag.Parse()

	config := server.ParseConfig(*config_file)

	res_index_data, err := ioutil.ReadFile(config.IndexPath)
	if err != nil {
		log.Fatal(err)
	}

	var res_index interface{}

	err = json.Unmarshal(res_index_data, &res_index)
	if err != nil {
		log.Fatal(err)
	}

	//list of valid game character sprite resource keys
	var spriteNames []string
	for k := range res_index.(map[string]interface{})["cache"].(map[string]interface{})["charset"].(map[string]interface{}) {
		if k != "_dirname" {
			spriteNames = append(spriteNames, k)
		}
	}

	//list of valid sound resource keys
	var soundNames []string
	for k := range res_index.(map[string]interface{})["cache"].(map[string]interface{})["sound"].(map[string]interface{}) {
		if k != "_dirname" {
			soundNames = append(soundNames, k)
		}
	}

	//list of valid system resource keys
	var systemNames []string
	for k := range res_index.(map[string]interface{})["cache"].(map[string]interface{})["system"].(map[string]interface{}) {
		if k != "_dirname" {
			systemNames = append(systemNames, k)
		}
	}
	
	//list of invalid sound names
	var ignoredSoundNames []string
	if config.BadSounds != "" {
		ignoredSoundNames = strings.Split(config.BadSounds, ",")
	}
	
	//list of valid picture names
	var pictureNames []string
	if config.PictureNames != "" {
		pictureNames = strings.Split(config.PictureNames, ",")
	}
	
	// list of valid picture prefixes
	var picturePrefixes []string
	if config.PicturePrefixes != "" {
		picturePrefixes = strings.Split(config.PicturePrefixes, ",")
	}

	var roomNames []string
	badRooms := strings.Split(config.BadRooms, ",")

	for i:=0; i < config.NumRooms; i++ {
		if !contains(badRooms, i) {
			roomNames = append(roomNames, strconv.Itoa(i))
		}
	}

	server.CreateAllHubs(roomNames, spriteNames, systemNames, soundNames, ignoredSoundNames, pictureNames, picturePrefixes, config.GameName, config.BlockIPs == 1)

	log.SetOutput(&lumberjack.Logger{
		Filename:   config.Logging.File,
		MaxSize:    config.Logging.MaxSize,
		MaxBackups: config.Logging.MaxBackups,
		MaxAge:     config.Logging.MaxAge,
	})
	log.SetFlags(log.Ldate | log.Ltime)

	http.Handle("/", http.FileServer(http.Dir("public/")))
	http.HandleFunc("/players", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strconv.Itoa(server.GetPlayerCount()))
	})
	log.Fatalf("%v %v \"%v\" %v", config.IP, "server", http.ListenAndServe(":" + strconv.Itoa(config.Port), nil), 500)
}
