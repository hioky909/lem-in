package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time" // Import du package time
)

// Structure pour représenter une salle
type Salle struct {
	nom        string
	x          int
	y          int
	voisins    []*Salle
	marquee    bool
	dansChemin bool
	index      int
	nodeIn     *Node
	nodeOut    *Node
}

// Structure pour représenter une arête dans le graphe de flux
type Edge struct {
	to       *Node
	capacity int
	flow     int
	reverse  *Edge
}

// Structure pour représenter un nœud dans le graphe de flux
type Node struct {
	name  string
	edges []*Edge
}

// Structure pour représenter un chemin
type Chemin struct {
	salles   []*Salle
	longueur int
}

// Structure pour représenter une fourmi
type Fourmi struct {
	numero   int
	position int
	chemin   *Chemin
}

var (
	salles       = make(map[string]*Salle)
	listeSalles  []*Salle
	listeChemins []*Chemin
	nbFourmis    int
	salleDepart  *Salle
	salleArrivee *Salle
	nodes        = make(map[string]*Node)
	startNode    *Node
	endNode      *Node
)

// Fonction principale
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . [fichier]")
		return
	}

	// Lecture et analyse du fichier d'entrée
	err := lireFichier(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Affichage du contenu du fichier
	afficherEntree()

	// Exécution de l'algorithme d'Edmonds-Karp pour trouver le flux maximum
	maxFlow := edmondsKarp(startNode, endNode)
	if maxFlow == 0 {
		fmt.Println("ERROR: Pas de chemin entre la salle de départ et d'arrivée.")
		return
	}

	// Extraction des chemins utilisés dans le flux maximum
	listeChemins = extraireChemins()
	if len(listeChemins) == 0 {
		fmt.Println("ERROR: Pas de chemin trouvé après extraction.")
		return
	}

	// Attribution des fourmis aux chemins
	attribuerFourmis()

	// Simulation du déplacement des fourmis avec chronomètre et compteur supplémentaire
	simulerDeplacements()
}

// Fonction pour lire et analyser le fichier d'entrée
func lireFichier(nomFichier string) error {
	fichier, err := os.Open(nomFichier)
	if err != nil {
		return fmt.Errorf("ERROR: Impossible de lire le fichier.")
	}
	defer fichier.Close()

	scanner := bufio.NewScanner(fichier)
	etat := "ants" // État de lecture: ants, rooms, links
	ligneNum := 0
	var indexSalle int
	for scanner.Scan() {
		ligne := scanner.Text()
		ligneNum++

		// Ignorer les commentaires
		if strings.HasPrefix(ligne, "#") && !strings.HasPrefix(ligne, "##") {
			continue
		}

		if etat == "ants" {
			// Lecture du nombre de fourmis
			nb, err := strconv.Atoi(ligne)
			if err != nil || nb <= 0 {
				return fmt.Errorf("ERROR: Nombre de fourmis invalide à la ligne %d.", ligneNum)
			}
			nbFourmis = nb
			etat = "rooms"
			continue
		}

		// Gestion des commandes ##start et ##end
		if ligne == "##start" {
			etat = "start"
			continue
		}
		if ligne == "##end" {
			etat = "end"
			continue
		}

		// Gestion des salles
		if etat == "rooms" || etat == "start" || etat == "end" {
			if strings.Contains(ligne, "-") {
				etat = "links"
			} else {
				parts := strings.Fields(ligne)
				if len(parts) != 3 {
					return fmt.Errorf("ERROR: Format de salle invalide à la ligne %d.", ligneNum)
				}
				nom := parts[0]
				if strings.HasPrefix(nom, "L") || strings.HasPrefix(nom, "#") {
					return fmt.Errorf("ERROR: Nom de salle invalide à la ligne %d.", ligneNum)
				}
				x, err1 := strconv.Atoi(parts[1])
				y, err2 := strconv.Atoi(parts[2])
				if err1 != nil || err2 != nil {
					return fmt.Errorf("ERROR: Coordonnées invalides à la ligne %d.", ligneNum)
				}
				if _, existe := salles[nom]; existe {
					return fmt.Errorf("ERROR: Salle en double à la ligne %d.", ligneNum)
				}
				// Créer la salle
				salle := &Salle{
					nom:     nom,
					x:       x,
					y:       y,
					voisins: []*Salle{},
					index:   indexSalle,
				}
				indexSalle++
				salles[nom] = salle
				listeSalles = append(listeSalles, salle)

				// Créer des nœuds pour le graphe de flux
				nodeIn := &Node{name: nom + "_in"}
				nodeOut := &Node{name: nom + "_out"}
				nodes[nom+"_in"] = nodeIn
				nodes[nom+"_out"] = nodeOut
				salle.nodeIn = nodeIn
				salle.nodeOut = nodeOut

				// Capacité entre nodeIn et nodeOut
				capacity := 1
				if etat == "start" || etat == "end" {
					capacity = nbFourmis // Capacité égale au nombre de fourmis pour start et end
				}
				edge := &Edge{to: nodeOut, capacity: capacity}
				reverseEdge := &Edge{to: nodeIn, capacity: 0}
				edge.reverse = reverseEdge
				reverseEdge.reverse = edge
				nodeIn.edges = append(nodeIn.edges, edge)
				nodeOut.edges = append(nodeOut.edges, reverseEdge)

				if etat == "start" {
					salleDepart = salle
					startNode = nodeOut // Le nœud de sortie du start
					etat = "rooms"
				}
				if etat == "end" {
					salleArrivee = salle
					endNode = nodeIn // Le nœud d'entrée du end
					etat = "rooms"
				}
				continue
			}
		}

		// Gestion des liens
		if etat == "links" {
			parts := strings.Split(ligne, "-")
			if len(parts) != 2 {
				return fmt.Errorf("ERROR: Format de lien invalide à la ligne %d.", ligneNum)
			}
			salle1, existe1 := salles[parts[0]]
			salle2, existe2 := salles[parts[1]]
			if !existe1 || !existe2 {
				return fmt.Errorf("ERROR: Lien vers une salle inconnue à la ligne %d.", ligneNum)
			}
			// Vérifier si le lien n'existe pas déjà
			if !contientSalle(salle1.voisins, salle2) {
				salle1.voisins = append(salle1.voisins, salle2)
				salle2.voisins = append(salle2.voisins, salle1)
			}

			// Ajouter les arêtes au graphe de flux
			ajouterArete(salle1.nodeOut, salle2.nodeIn)
			ajouterArete(salle2.nodeOut, salle1.nodeIn)
			continue
		}
	}

	if salleDepart == nil || salleArrivee == nil {
		return fmt.Errorf("ERROR: Salle de départ ou d'arrivée manquante.")
	}

	return nil
}

// Fonction pour vérifier si une liste de salles contient une salle donnée
func contientSalle(salles []*Salle, salle *Salle) bool {
	for _, s := range salles {
		if s == salle {
			return true
		}
	}
	return false
}

// Fonction pour ajouter une arête entre deux nœuds
func ajouterArete(from, to *Node) {
	if from == nil || to == nil {
		return
	}
	edge := &Edge{to: to, capacity: 1}
	reverseEdge := &Edge{to: from, capacity: 0}
	edge.reverse = reverseEdge
	reverseEdge.reverse = edge
	from.edges = append(from.edges, edge)
	to.edges = append(to.edges, reverseEdge)
}

// Fonction pour afficher le contenu de l'entrée
func afficherEntree() {
	fmt.Println(nbFourmis)
	for _, salle := range listeSalles {
		if salle == salleDepart {
			fmt.Println("##start")
		}
		if salle == salleArrivee {
			fmt.Println("##end")
		}
		fmt.Printf("%s %d %d\n", salle.nom, salle.x, salle.y)
	}
	for _, salle := range listeSalles {
		for _, voisin := range salle.voisins {
			if salle.index < voisin.index {
				fmt.Printf("%s-%s\n", salle.nom, voisin.nom)
			}
		}
	}
}

// Fonction pour exécuter l'algorithme d'Edmonds-Karp
func edmondsKarp(source, sink *Node) int {
	maxFlow := 0
	for {
		parent := make(map[*Node]*Node)
		edgeTo := make(map[*Node]*Edge)
		// BFS
		queue := []*Node{source}
		parent[source] = nil
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			for _, edge := range u.edges {
				residualCapacity := edge.capacity - edge.flow
				if residualCapacity > 0 {
					if _, found := parent[edge.to]; !found && edge.to != source {
						parent[edge.to] = u
						edgeTo[edge.to] = edge
						if edge.to == sink {
							break
						}
						queue = append(queue, edge.to)
					}
				}
			}
			if _, found := parent[sink]; found {
				break
			}
		}
		if _, found := parent[sink]; !found {
			break
		}
		// Trouver le flux maximum qu'on peut ajouter
		pathFlow := int(^uint(0) >> 1) // Max int
		for v := sink; v != source; v = parent[v] {
			edge := edgeTo[v]
			residualCapacity := edge.capacity - edge.flow
			if residualCapacity < pathFlow {
				pathFlow = residualCapacity
			}
		}
		// Mettre à jour les flux
		for v := sink; v != source; v = parent[v] {
			edge := edgeTo[v]
			edge.flow += pathFlow
			edge.reverse.flow -= pathFlow
		}
		maxFlow += pathFlow
	}
	return maxFlow
}

// Fonction pour extraire les chemins du flux maximum
func extraireChemins() []*Chemin {
	chemins := []*Chemin{}
	usedEdges := make(map[*Edge]bool)

	for {
		node := startNode
		visited := make(map[*Node]bool)
		visited[node] = true
		edgeTo := make(map[*Node]*Edge)
		// BFS pour trouver un chemin avec flow > 0
		queue := []*Node{node}
		found := false
		for len(queue) > 0 && !found {
			u := queue[0]
			queue = queue[1:]
			for _, edge := range u.edges {
				if edge.flow > 0 && !visited[edge.to] && !usedEdges[edge] {
					visited[edge.to] = true
					edgeTo[edge.to] = edge
					if edge.to == endNode {
						found = true
						break
					}
					queue = append(queue, edge.to)
				}
			}
		}
		if !found {
			break
		}
		// Reconstruire le chemin
		pathNodes := []*Node{endNode}
		for v := endNode; v != startNode; v = edgeTo[v].reverse.to {
			pathNodes = append([]*Node{edgeTo[v].reverse.to}, pathNodes...)
			// Décrémenter le flux de l'arête
			edgeTo[v].flow -= 1
			edgeTo[v].reverse.flow += 1
			usedEdges[edgeTo[v]] = true
		}
		// Convertir pathNodes en chemin de *Salle
		pathSalles := []*Salle{}
		visitedSalles := make(map[*Salle]bool)
		for _, node := range pathNodes {
			// Ignorer les nœuds internes (_in et _out)
			roomName := strings.TrimSuffix(strings.TrimSuffix(node.name, "_in"), "_out")
			salle, existe := salles[roomName]
			if !existe {
				continue
			}
			if !visitedSalles[salle] {
				pathSalles = append(pathSalles, salle)
				visitedSalles[salle] = true
			}
		}
		chemin := &Chemin{
			salles:   pathSalles,
			longueur: len(pathSalles),
		}
		chemins = append(chemins, chemin)
	}

	return chemins
}

// Fonction pour attribuer les fourmis aux chemins
func attribuerFourmis() {
	// Trier les chemins par longueur croissante
	sort.Slice(listeChemins, func(i, j int) bool {
		return listeChemins[i].longueur < listeChemins[j].longueur
	})
}

// Fonction pour réinitialiser les marques des salles
func resetMarques() {
	for _, salle := range listeSalles {
		salle.marquee = false
	}
}

// Fonction pour simuler les déplacements des fourmis avec chronomètre et compteur supplémentaire
func simulerDeplacements() {
	// Calculer le nombre de fourmis à attribuer à chaque chemin
	nombreChemins := len(listeChemins)
	fourmisRestantes := nbFourmis
	attributionFourmis := make([]int, nombreChemins)

	// Attribution des fourmis aux chemins en minimisant le nombre de tours
	for fourmisRestantes > 0 {
		meilleurIndex := -1
		minTempsTotal := int(^uint(0) >> 1) // Max int
		for i := 0; i < nombreChemins; i++ {
			tempsTotal := (listeChemins[i].longueur - 1) + attributionFourmis[i]
			if tempsTotal < minTempsTotal {
				minTempsTotal = tempsTotal
				meilleurIndex = i
			}
		}
		attributionFourmis[meilleurIndex]++
		fourmisRestantes--
	}

	// Initialiser les fourmis
	numeroFourmi := 1
	fileFourmis := []*Fourmi{}
	for i, nb := range attributionFourmis {
		for j := 0; j < nb; j++ {
			fourmi := &Fourmi{
				numero:   numeroFourmi,
				position: 0,
				chemin:   listeChemins[i],
			}
			numeroFourmi++
			fileFourmis = append(fileFourmis, fourmi)
		}
	}

	// Simulation des déplacements avec compteur de tours et chronomètre
	termine := false
	tours := 0              // Initialisation du compteur de tours
	antsAtteintes := 0      // Compteur pour les fourmis atteignant la salle d'arrivée
	startTime := time.Now() // Enregistrement du temps de début

	for !termine {
		termine = true
		sortie := ""
		for _, fourmi := range fileFourmis {
			if fourmi.position < len(fourmi.chemin.salles)-1 {
				// Vérifier si la prochaine salle est libre ou est la salle d'arrivée
				prochaineSalle := fourmi.chemin.salles[fourmi.position+1]
				if prochaineSalle == salleArrivee || !prochaineSalle.marquee {
					// Déplacer la fourmi
					if prochaineSalle != salleArrivee {
						prochaineSalle.marquee = true
					}
					actuelleSalle := fourmi.chemin.salles[fourmi.position]
					if actuelleSalle != salleDepart {
						actuelleSalle.marquee = false
					}
					fourmi.position++
					if prochaineSalle == salleArrivee {
						antsAtteintes++ // Incrémenter le compteur de fourmis atteintes
					}
					if sortie != "" {
						sortie += " "
					}
					sortie += fmt.Sprintf("L%d-%s", fourmi.numero, prochaineSalle.nom)
					termine = false
				}
			}
		}
		if sortie != "" {
			fmt.Println(sortie)
			tours++ // Incrémenter le compteur de tours
		}
	}

	endTime := time.Now()              // Enregistrement du temps de fin
	duration := endTime.Sub(startTime) // Calcul de la durée totale

	// Afficher les résultats
	fmt.Printf("Nombre de tours : %d\n", tours)
	fmt.Printf("Nombre de fourmis atteintes : %d\n", antsAtteintes)
	fmt.Printf("Temps d'exécution : %v\n", duration)
}
