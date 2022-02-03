package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

type DataDragon struct {
	Type    string
	Format  string
	Version string
	Data    map[string]RawChampion
}

type RawChampion struct {
	ID    string
	Name  string
	Title string
	Lore  string
	Skins []map[string]interface{}
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalln("You must pass two versions as arguments")
		return
	}

	version := os.Args[1]
	log.Println("Latest version: ", version)
	log.Println("Previous version: ", os.Args[2])

	err := os.MkdirAll(filepath.Join("public", "data"), 0777)

	if err != nil {
		log.Panicln(err)
	}

	latest := getPatchChampions(version)
	previous := getPatchChampions(os.Args[2])

	keys := make([]string, len(latest))
	i := 0

	for k := range latest {
		keys[i] = k
		i++
	}

	sort.Strings(keys)
	champions := make([]Champion, len(keys))

	for i, key := range keys {
		_, ok := previous[key]
		champions[i] = getChampion(key, version, !ok)
		log.Println(key)
	}

	patch := struct {
		Version   string     `json:"version"`
		Champions []Champion `json:"champions"`
	}{version, champions}

	fileData, err := json.MarshalIndent(patch, "", " ")

	if err != nil {
		log.Panicln(err)
	}

	err = os.WriteFile(filepath.Join("public", "data", "patch.json"), fileData, 0644)

	if err != nil {
		log.Panicln(err)
	}

	err = os.WriteFile(filepath.Join("public", "data", "version.txt"), []byte(version), 0644)

	if err != nil {
		log.Panicln(err)
	}

	log.Printf("Patch %s successfully downloaded", version)
}

func getPatchChampions(version string) map[string]RawChampion {
	resp, err := http.Get(fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", version))

	if err != nil {
		log.Panicln(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Panicln(err)
	}

	var dragon DataDragon
	err = json.Unmarshal(body, &dragon)

	if err != nil {
		log.Panicln(err)
	}

	return dragon.Data
}

type Champion struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Lore      string `json:"lore"`
	Thumbnail string `json:"thumbnail"`
	Skins     []Skin `json:"skins"`
	New       bool   `json:"new"`
}

type Skin struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func getChampion(id string, version string, isNew bool) Champion {
	resp, err := http.Get(fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion/%s.json", version, id))

	if err != nil {
		log.Panicln(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Panicln(err)
	}

	var dragon DataDragon
	err = json.Unmarshal(body, &dragon)

	if err != nil {
		log.Panicln(err)
	}

	raw := dragon.Data[id]

	champion := Champion{
		ID:        raw.ID,
		Name:      raw.Name,
		Title:     raw.Title,
		Lore:      raw.Lore,
		Thumbnail: fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/img/champion/loading/%s_0.jpg", id),
		Skins:     getSkins(id, raw.Skins),
		New:       isNew,
	}

	return champion
}

func getSkins(id string, rawSkins []map[string]interface{}) []Skin {
	skins := make([]Skin, len(rawSkins))

	for i, value := range rawSkins {
		skins[i] = Skin{
			Name: fmt.Sprintf("%v", value["name"]),
			Url:  fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/img/champion/splash/%v_%v.jpg", id, value["num"]),
		}
	}

	return skins
}
