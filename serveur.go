package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"sync"
)

var w sync.WaitGroup

func main() {
	var ipAdress string
	flag.StringVar(&ipAdress, "ip", "127.0.0.1", "Rentrer l'ip du serveur")
	flag.Parse()
	var nbConnection int
	listener, err := net.Listen("tcp", ipAdress+":8080")
	if err != nil {
		log.Println("listen error:", err)
		return
	}
	defer listener.Close()
	var conns = make([]net.Conn, 4)

	for nbConnection != 4 {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			return
		} else {
			conns[nbConnection] = conn
			numeroConn := []byte(strconv.Itoa(nbConnection))
			_, err := conn.Write(numeroConn)
			if err != nil {
				log.Println("write error numeroConn n°", nbConnection, " : ", err)
				return
			}
			for i := 0; i < nbConnection; i++ {
				_, err := conns[i].Write(numeroConn)
				if err != nil {
					log.Println("write error2 numeroConn n°", nbConnection, " : ", err)
					return
				}
			}
		}
		nbConnection++
		defer conn.Close()
		log.Println("Un client s'est connecté")
	}

	log.Println("Tous le monde est connecté")

	msgTousConnecte := []byte{'1'}
	for _, conn := range conns {
		_, err := conn.Write(msgTousConnecte)
		if err != nil {
			log.Println("write error msgTousConnecte:", err)
			return
		}
	}

	cPersos := make([]chan string, 4)
	for i := 0; i < 4; i++ {
		cPersos[i] = make(chan string, 1)
		go perso(conns[i], i, cPersos[i])
		w.Add(1)
	}

	w.Wait() // attent que tous les joueurs choisissent leur perso

	log.Println("Tous le monde a choisi son personnage")

	var tousPerso string
	for i := 0; i < 3; i++ {
		tousPerso += <-cPersos[i]
		tousPerso += ","
	}
	tousPerso += <-cPersos[3]

	msgTousPersoChoisi := []byte(tousPerso)
	for _, conn := range conns {
		_, err := conn.Write(msgTousPersoChoisi)
		if err != nil {
			log.Println("write error msgTousPersoChoisi:", err)
			return
		}
	}

	//refaire en boucle
	for {
		cScore := make([]chan string, 4)
		for i := 0; i < 4; i++ {
			cScore[i] = make(chan string, 1)
			go score(conns[i], cScore[i], i)
			w.Add(1)
		}

		w.Wait() // attent que tous les joueurs finissent leur course et l'envoie de leur temps

		var tempsTousJoueurs string
		for i := 0; i < 3; i++ {
			tempsTousJoueurs += <-cScore[i]
			tempsTousJoueurs += ","
		}
		tempsTousJoueurs += <-cScore[3]

		msgTemps := []byte(tempsTousJoueurs)
		for _, conn := range conns {
			_, err := conn.Write(msgTemps)
			if err != nil {
				log.Println("write error msgTemps:", err)
				return
			}
		}

		for i := 0; i < 4; i++ {
			go rejoue(conns[i], i)
			w.Add(1)
		}

		w.Wait()

		msgLanceNouvellePartie := []byte{'5'}
		for _, conn := range conns {
			_, err := conn.Write(msgLanceNouvellePartie)
			if err != nil {
				log.Println("write error msgLanceNouvellePartie:", err)
				return
			}
		}
	}
}

func perso(conn net.Conn, nb int, cPersos chan string) {
	msgPersoChoisi := make([]byte, 1)
	_, err := conn.Read(msgPersoChoisi)
	if err != nil {
		log.Println("Erreur msg choix perso :", err)
		return
	}
	log.Println("Le joueur ", nb, " a choisi son personnage")
	cPersos <- string(msgPersoChoisi)
	w.Done()
}

func score(conn net.Conn, cScore chan string, nb int) {
	msgScore := make([]byte, 10)
	_, err := conn.Read(msgScore)
	if err != nil {
		log.Println("Erreur msg score :", err)
		return
	} else {
		log.Println("Le joueur ", nb, " a fini sa course et son temps est de ", string(msgScore), " millisecondes")
	}
	cScore <- string(msgScore)
	w.Done()
	return
}

func rejoue(conn net.Conn, nb int) {
	msgRejouer := make([]byte, 1)
	_, err := conn.Read(msgRejouer)
	if err != nil {
		log.Println("Erreur msg rejouer :", err)
		return
	}
	if "4" == string(msgRejouer) {
		log.Println("Le joueur ", nb, " veut rejouer")
	}
	w.Done()

}
