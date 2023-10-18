package v2

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
)

// Get the list of versions, select the latest one and download the champions.
// The champions are saved in the public/data/patch.json file.
func GetLatestVersion() {
	fileName := filepath.Join("public", "data", "version.txt")

	// Check if the latest version is already downloaded
	_, err := os.Stat(fileName)

	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}

		err = os.MkdirAll(filepath.Dir(fileName), 0750)

		if err != nil {
			panic(err)
		}

		_, err = os.Create(fileName)

		if err != nil {
			panic(err)
		}
	}

	// Read the latest version from the file
	// and compare it with the latest version
	file, err := os.Open(fileName)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	versions := getVersions()
	currentVersion := scanner.Text()
	force := slices.Contains(os.Args, "--force")

	if currentVersion == versions[0] && !force {
		fmt.Println("Version", currentVersion, "is already downloaded. Use the flag --force to download it again")
		return
	}

	latestChampions := getChampionsByVersion(versions[0])
	previousChampions := getChampionsByVersion(versions[1])

	// Sort the champions by name
	keys := make([]string, len(latestChampions))
	champions := make([]Champion, len(keys))
	i := 0

	for k := range latestChampions {
		keys[i] = k
		i++
	}

	slices.Sort(keys)

	for i, key := range keys {
		_, isOld := previousChampions[key]
		champions[i] = getChampionByVersion(key, versions[0], !isOld)
		fmt.Println(key)
	}

	patch := struct {
		Version   string     `json:"version"`
		Champions []Champion `json:"champions"`
	}{versions[0], champions}

	data, err := json.MarshalIndent(patch, "", " ")

	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join("public", "data", "patch.json"), data, 0644)

	if err != nil {
		panic(err)
	}

	fmt.Println("Version", versions[0], "successfully downloaded")
}

// Get the list of versions
func getVersions() []string {
	resp, err := http.Get("https://ddragon.leagueoflegends.com/api/versions.json")

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var versions []string
	err = json.Unmarshal(body, &versions)

	if err != nil {
		panic(err)
	}

	return versions
}

// Get the id of the all champions of a specific version
func getChampionsByVersion(version string) map[string]dragonChampion {
	resp, err := http.Get(fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", version))

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var dragon struct {
		Data map[string]dragonChampion
	}

	err = json.Unmarshal(body, &dragon)

	if err != nil {
		panic(err)
	}

	return dragon.Data
}

// Struct of the champion of the data dragon response
type dragonChampion struct {
	Id    string
	Name  string
	Title string
	Lore  string
	Skins []dragonSkin
}

type dragonSkin struct {
	Id   string
	Num  int
	Name string
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

// Get the details of a champion.
func getChampionByVersion(id string, version string, isNew bool) Champion {
	resp, err := http.Get(fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion/%s.json", version, id))

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var dragon struct {
		Data map[string]dragonChampion
	}

	err = json.Unmarshal(body, &dragon)

	if err != nil {
		panic(err)
	}

	raw := dragon.Data[id]

	champion := Champion{
		ID:        raw.Id,
		Name:      raw.Name,
		Title:     raw.Title,
		Lore:      raw.Lore,
		Thumbnail: fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/img/champion/loading/%s_0.jpg", id),
		Skins:     getSkins(id, raw.Skins),
		New:       isNew,
	}

	return champion
}

// Get the skins of a champion
func getSkins(id string, dragonSkins []dragonSkin) []Skin {
	skins := make([]Skin, len(dragonSkins))

	for i, value := range dragonSkins {
		skins[i] = Skin{
			Name: fmt.Sprintf("%v", value.Name),
			Url:  fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/img/champion/splash/%v_%v.jpg", id, value.Num),
		}
	}

	return skins
}
