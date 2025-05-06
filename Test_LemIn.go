package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	// Liste des fichiers de test à exécuter
	testFiles := []string{
		"example00.txt",
		"example01.txt",
		"example02.txt",
		"example03.txt",
		"example04.txt",
		"example05.txt",
		"badexample00",
		"badexample01",
		"example06.txt",
		"example07.txt",
	}

	total := len(testFiles)
	passed := 0
	failed := 0

	fmt.Println("Début de l'exécution des tests...\n")

	// Itérer sur chaque fichier de test
	for _, testFile := range testFiles {
		fmt.Printf("Exécution du test : %s\n", testFile)

		// Préparer la commande à exécuter
		cmd := exec.Command("go", "run", "LemIn.go", testFile)

		// Exécuter la commande et capturer la sortie (stdout et stderr)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Afficher la sortie du programme
		fmt.Println("Sortie du programme :")
		fmt.Println(strings.TrimSpace(outputStr))

		// Vérifier si la commande s'est exécutée sans erreur
		if err != nil {
			// Commande échouée
			fmt.Printf("Résultat : ÉCHEC (Erreur : %v)\n\n", err)
			failed++
		} else {
			// Commande réussie
			fmt.Printf("Résultat : RÉUSSI\n\n")
			passed++
		}
	}

	// Afficher le résumé des tests
	fmt.Println("Résumé des Tests :")
	fmt.Printf("Nombre total de tests : %d\n", total)
	fmt.Printf("Tests réussis : %d\n", passed)
	fmt.Printf("Tests échoués : %d\n", failed)

	// Optionnel : Définir le code de sortie basé sur les résultats des tests
	if failed > 0 {
		fmt.Println("\nCertains tests ont échoué.")
		// Vous pouvez décommenter la ligne suivante pour que le programme retourne un code d'erreur
		// os.Exit(1)
	} else {
		fmt.Println("\nTous les tests ont réussi.")
		// Vous pouvez décommenter la ligne suivante pour que le programme retourne un code de succès
		// os.Exit(0)
	}
}
