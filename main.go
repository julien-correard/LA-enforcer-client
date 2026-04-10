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

	err = postHighscore(hs, name)
	if err != nil {
		return fmt.Errorf("échec de l'envoi du score")
	}

	fmt.Println("Score envoyé avec succès")
	fmt.Println("Nom :", name)
	fmt.Println("Score :", hs.Score)
	fmt.Println("Date :", hs.Date)

	lines[3] = "true"
	writeFile(lines)

	return nil
}

func postHighscore(hs Highscore, name string) error {
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
		"https://la-enforcer-server.onrender.com/scores",
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
	fmt.Println("\nAppuyez sur Entrée pour quitter...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
