# neurau.eu

Site web de la communauté francophone Neurau (psychologie moderne).
Généré avec Hugo, éditable via Decap CMS, hébergé sur VPS.
Commentaires via Isso.

## Développement local

Prérequis : Hugo extended >= 0.128

    make serve

Le site est accessible sur http://localhost:1313/

Pour générer le site statique :

    make build

Pour compiler le service CMS (OAuth + webhook) :

    make cms

## Structure

    content/          Contenu Markdown (blog, connaissances, pages)
    layouts/          Templates Hugo
    assets/css/       Feuilles de style (CSS vanilla, variables custom)
    assets/js/        JavaScript (menu mobile, dark mode)
    static/admin/     Interface Decap CMS
    static/images/    Images du site
    config/_default/  Configuration Hugo
    deploy/cms/       Service OAuth + webhook (Go)
    deploy/isso/      Déploiement Isso (commentaires)
    GNUmakefile       Commandes communes

## Déploiement sur VPS

### 1. DNS

Chez le registrar, ajouter les enregistrements `A` (et `AAAA` si
IPv6) pointant `neurau.eu` et `comments.neurau.eu` vers l'IP du VPS.

### 2. GitHub OAuth App

Créer une OAuth App dans GitHub (Settings > Developer settings >
OAuth Apps) :

- Application name : Neurau CMS
- Homepage URL : https://neurau.eu
- Authorization callback URL : https://neurau.eu/oauth/callback

Noter le Client ID et le Client Secret.

### 3. GitHub Webhook

Dans le repo GitHub (Settings > Webhooks), ajouter un webhook :

- Payload URL : https://neurau.eu/webhook
- Content type : application/json
- Secret : choisir un secret aléatoire
- Events : Just the push event

### 4. Installation

Cloner le repo sur le VPS dans `/var/www/neurau.eu` puis créer le
fichier de secrets :

    cp deploy/cms/env.example deploy/cms/env
    # éditer deploy/cms/env avec les vrais secrets GitHub
    # éditer deploy/isso/secrets.cfg avec le mot de passe SMTP

Lancer l'installation complète :

    sudo make deploy

### Targets disponibles

    make build            # générer le site statique
    make serve            # serveur de développement local
    make cms              # compiler le service CMS
    make clean            # nettoyer les fichiers générés
    sudo make deploy      # installation complète sur VPS
    sudo make deploy-isso # (re)configurer le conteneur Isso
    sudo make deploy-cms  # (re)installer le service CMS
    sudo make deploy-nginx   # (re)configurer nginx
    sudo make deploy-certbot # (re)configurer certbot

## Decap CMS

L'interface d'édition est accessible sur https://neurau.eu/admin/

L'authentification passe par GitHub OAuth. Seuls les utilisateurs
avec accès au repo peuvent se connecter et éditer le contenu. Chaque
modification crée un commit dans le repo GitHub, ce qui déclenche
le webhook et un rebuild automatique du site sur le VPS.

## Commentaires (Isso)

Les commentaires sont activés sur les articles de blog. Isso tourne
sur le VPS dans un conteneur Podman, derrière nginx en reverse proxy
sur `comments.neurau.eu`.

Adapter les paramètres SMTP dans `deploy/isso/isso.cfg` pour la
notification des nouveaux commentaires par email.

### Sécurité

Le reverse proxy nginx bloque toute requête dont l'origin n'est pas
`https://neurau.eu` (CORS strict + retour 403). Le rate limiting est
configuré à 5 requêtes/seconde par IP. La modération des commentaires
est activée par défaut.

L'accès direct à `comments.neurau.eu` depuis un navigateur renvoie
une erreur 403 car l'en-tête `Origin` ne correspond pas.

## Contenu

### Blog

Ajouter des articles dans `content/blog/`. Chaque article a un front
matter avec titre, date, description, catégories et tags.

### Base de connaissances

Organiser les articles par sous-dossiers dans `content/connaissances/`.
Chaque sous-dossier représente une catégorie et contient un `_index.md`
avec le titre et la description de la catégorie.

### Pages

Les pages statiques (accueil, à propos, contact) sont à la racine de
`content/`.
