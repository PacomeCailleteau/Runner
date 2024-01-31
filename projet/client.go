package main

import (
	"log"
	"net"
	"strconv"
	"strings"
)

func connection(ipAdress string, g *Game) {

	var placement int

	conn, err := net.Dial("tcp", ipAdress+":8080")
	if err != nil {
		log.Println("Dial error:", err)
		return
	}
	defer conn.Close()
	msgPlacement := make([]byte, 1)

	_, err = conn.Read(msgPlacement)
	if err != nil {
		log.Println("Erreur :", err)
		return
	}
	placement, _ = strconv.Atoi(string(msgPlacement))
	g.pos = placement
	log.Println("Je suis connecté à la position ", placement, ".")
	g.cNbJoueurCo <- placement

	msgNbJoueur := make([]byte, 1)
	ok := placement == 3
	for !ok {
		_, err = conn.Read(msgNbJoueur)
		if err != nil {
			log.Println("Erreur explicite :", err)
			return
		}
		nb, _ := strconv.Atoi(string(msgNbJoueur))
		ok = nb == 3
		<-g.cNbJoueurCo
		g.cNbJoueurCo <- nb
	}

	msgTousConnecte := make([]byte, 1)

	_, err = conn.Read(msgTousConnecte)
	if err != nil {
		log.Println("Erreur :", err)
		return
	}

	log.Println("Tout le monde est connecté.")
	g.cEcritureClient <- true

	g.cOk <- true
	var persoChoisi = false
	for !persoChoisi {
		select {
		case <-g.cTemp:
			persoChoisi = true
		default:
		}
	}

	log.Println("Personnage choisi")

	couleur := strconv.Itoa(g.runners[placement].colorScheme)
	msgPersoChoisi := []byte(couleur)
	_, err = conn.Write(msgPersoChoisi)
	if err != nil {
		log.Println("Erreur :", err)
		return
	}

	msgTousPersoChoisi := make([]byte, 7)

	_, err = conn.Read(msgTousPersoChoisi)
	if err != nil {
		log.Println("Erreur :", err)
		return
	}

	tabPerso := strings.Split(string(msgTousPersoChoisi), ",")
	for i := range tabPerso {
		p, _ := strconv.Atoi(tabPerso[i])
		g.runnersColors[i] = p
	}

	log.Println("Tout le monde a choisi son personnage.")
	g.cEcritureClient <- true

	for {
		<-g.cTemp

		temps := int(g.runners[placement].runTime.Milliseconds())
		msgCourseFinie := []byte(strconv.Itoa(temps))
		_, err = conn.Write(msgCourseFinie)
		if err != nil {
			log.Println("Erreur :", err)
			return
		}

		msgTousFini := make([]byte, 50)

		_, err = conn.Read(msgTousFini)
		if err != nil {
			log.Println("Erreur :", err)
			return
		}

		msgResult := string(msgTousFini)

		g.cEcritureClient <- true

		g.cResultat <- msgResult
		go rejouer(0, conn, g)
		var rejoue = false
		for !rejoue {
			select {
			case <-g.cTemp:
				rejoue = true
			default:
			}
		}
		msgTousRejoue := make([]byte, 1)

		_, err = conn.Read(msgTousRejoue)
		if err != nil {
			log.Println("Erreur message tous le monde rejoue :", err)
			return
		}
		g.cEcritureClient <- true
	}

}

func rejouer(nbRejoue int, conn net.Conn, g *Game) {
	ok := nbRejoue == 4
	msgNbJoueur := make([]byte, 1)
	g.crejoue <- 0
	for !ok {
		_, err := conn.Read(msgNbJoueur)
		if err != nil {
			log.Println("Erreur explicite :", err)
			return
		}
		nb, _ := strconv.Atoi(string(msgNbJoueur))
		ok = nb == 4
		<-g.crejoue
		g.crejoue <- nb
	}
	msgRejoue := []byte("4")
	_, err := conn.Write(msgRejoue)
	if err != nil {
		log.Println("Erreur envoi message rejoue :", err)
		return
	}
}
