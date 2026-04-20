package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const maxLenName = 15
const xorKey = 0x5A
const file = "SCORE.DAT"

type Highscore struct {
	Score       int
	Version     string
	Date        string
	AlreadySent bool
}

func main() {
	err := run()
	if err != nil {
		fmt.Println("Erreur:", err)
		waitForKey()
		os.Exit(1)
	}

	waitForKey()
}

func run() error {

	baseURL := getBaseURL()

	// On réveille le serveur
	go func() {
		client := &http.Client{
			Timeout: 1 * time.Second,
		}

		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}()

	fmt.Println("LA Enforcer score sender © 1992 Ironbyte studios")
	fmt.Println("Lecture du fichier score.dat")

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("fichier score.dat non trouvé")
	}

	xorCipher(data, xorKey)

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	if len(lines) < 4 {
		return fmt.Errorf("fichier score.dat corrompu")
	}

	header := lines[0]
	const expectedHeader = "la-enforcer-score-file"

	parts := strings.Split(header, "v")
	if len(parts) != 2 || parts[0] != expectedHeader {
		return fmt.Errorf("fichier score.dat invalide")
	}

	version := parts[1]

	score, err := strconv.Atoi(lines[1])
	if err != nil {
		return fmt.Errorf("score invalide")
	}

	alreadySent, err := strconv.ParseBool(lines[3])
	if err != nil {
		return fmt.Errorf("flag AlreadySent invalide")
	}

	hs := Highscore{
		Version:     version,
		Score:       score,
		Date:        lines[2],
		AlreadySent: alreadySent,
	}

	if hs.AlreadySent {
		fmt.Println("Score déjà envoyé.")
		return nil
	}

	name, err := askName()
	if err != nil {
		return err
	}

	err = waitForServerReady(baseURL+"/health", 240*time.Second)
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		err = postHighscore(baseURL, hs, name)
		if err == nil {
			break
		}
		if i < 2 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return fmt.Errorf("échec de l'envoi du score: %w", err)
	}

	fmt.Println("Score envoyé avec succès")
	fmt.Println("Nom :", name)
	fmt.Println("Score :", hs.Score)
	fmt.Println("Date :", hs.Date)

	lines[3] = "true"
	writeFile(lines)

	return nil
}

func waitForServerReady(url string, timeout time.Duration) error {

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	start := time.Now()
	fmt.Print("Connexion au serveur (le réveil peut être un peu lent)")

	for {
		// timeout global
		if time.Since(start) > timeout {
			return fmt.Errorf("serveur indisponible après %v", timeout)
		}

		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()

			if resp.StatusCode == 200 {
				fmt.Println(" OK")
				return nil
			}
		}

		// animation console rétro 😄
		fmt.Print(".")

		time.Sleep(2 * time.Second)
	}
}

func postHighscore(baseURL string, hs Highscore, name string) error {
	payload := map[string]interface{}{
		"score":       hs.Score,
		"date":        hs.Date,
		"playerName":  name,
		"gameVersion": hs.Version,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(
		baseURL+"/scores",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return nil
}

func askName() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Entrez votre nom: ")

		name, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		name = strings.TrimSpace(name)

		if len(name) == 0 || len(name) > maxLenName {
			fmt.Printf("Nom invalide (1 à %d caractères).\n", maxLenName)
			continue
		}

		return name, nil
	}
}

func writeFile(lines []string) {
	// 👇 reconstruire le contenu
	content := strings.Join(lines, "\n")

	data := []byte(content)

	// 👇 chiffrer
	xorCipher(data, xorKey)

	// 👇 écrire en binaire
	err := os.WriteFile(file, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func xorCipher(data []byte, key byte) {
	for i := range data {
		data[i] ^= key
	}
}

func waitForKey() {
	fmt.Println("\nLes scores sont consultables ici : https://julien-correard.github.io/LA-enforcer-web/")
	fmt.Println("\nAppuyez sur Entrée pour quitter...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func getBaseURL() string {
	url := os.Getenv("API_URL")
	if url == "" {
		// fallback (utile en dev)
		url = "https://la-enforcer-server.onrender.com"
	}
	return url
}
