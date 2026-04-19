# LA Enforcer – Client

## Présentation

Ce programme permet d'envoyer les scores du jeu rétro DOS *LA Enforcer* vers un serveur distant.

## Fonctionnalités

- Lecture du fichier `SCORE.DAT` généré par le jeu
- Dé-obfuscation des données via un chiffrement XOR
- Envoi du score au format JSON vers une API REST Spring Boot
- Prévention des envois multiples grâce à un flag local
- Attente active de la disponibilité du serveur (utile pour hébergement avec mise en veille)
- Réveil automatique du serveur via un appel HTTP en arrière-plan

## Détails techniques

Le fichier de score est volontairement obfusqué (XOR) afin de limiter les modifications triviales.  
Ce client se charge de la lecture, du décodage et de la communication HTTP avec le serveur.

Ce projet complète un système global comprenant :
- un jeu en C (génération des scores)
- une interface web en JavaScript (consultation des scores)
- un serveur Spring Boot (stockage des scores)

## Choix technologique

Ce projet est développé en Go pour sa simplicité et sa capacité à produire des exécutables multiplateformes.

Je me suis aidé d’outils d’intelligence artificielle comme support ponctuel lors du développement. Ce projet s’inscrit dans une démarche de reconversion, dans laquelle je me forme actuellement à plusieurs technologies (Java, HTML, CSS, PHP). Le langage Go est ici exploré de manière plus occasionnelle.

## Auteur

Julien Correard

## Licence

Ce projet est sous licence MIT. Voir le fichier LICENSE.
