# neurau.eu

Site web de la communauté francophone Neurau (psychologie moderne).
Généré avec Hugo, éditable via Decap CMS, hébergé sur Netlify.
Commentaires via Isso sur VPS.

## Développement local

Prérequis : Hugo extended >= 0.128

    make serve

Le site est accessible sur http://localhost:1313/

Pour générer le site statique :

    make build

## Structure

    content/          Contenu Markdown (blog, connaissances, pages)
    layouts/          Templates Hugo
    assets/css/       Feuilles de style (CSS vanilla, variables custom)
    assets/js/        JavaScript (menu mobile, dark mode)
    static/admin/     Interface Decap CMS
    static/images/    Images du site
    config/_default/  Configuration Hugo
    netlify.toml      Configuration Netlify (build, version Hugo)
    deploy/isso/      Déploiement Isso sur VPS (commentaires)
    GNUmakefile       Commandes communes

## Déploiement sur Netlify

### 1. Créer le site

- Créer un compte sur https://app.netlify.com/
- "Add new site" > "Import an existing project" > connecter le repo
  GitHub `rjarry/neurau.eu`
- Netlify détecte automatiquement `netlify.toml` et configure le build

### 2. Configurer le domaine

- Dans Site settings > Domain management > ajouter `neurau.eu`
- Chez le registrar DNS, créer un enregistrement :
  `CNAME neurau.eu -> <sous-domaine>.netlify.app`
- Netlify provisionne automatiquement le certificat TLS via Let's Encrypt

### 3. Activer Netlify Identity (pour Decap CMS)

- Site settings > Identity > Enable Identity
- Identity > Registration > Invite only
- Identity > Services > Git Gateway > Enable Git Gateway
- Inviter les éditeurs par email depuis l'onglet Identity

Les éditeurs invités pourront se connecter sur https://neurau.eu/admin/
et modifier le contenu directement depuis le navigateur.

## Decap CMS

L'interface d'édition est accessible sur https://neurau.eu/admin/

L'authentification passe par Netlify Identity. Seuls les utilisateurs
invités peuvent se connecter et éditer le contenu. Chaque modification
crée un commit dans le repo GitHub, ce qui déclenche automatiquement
un rebuild du site.

## Commentaires (Isso)

Les commentaires sont activés sur les articles de blog. Isso tourne sur
le VPS dans un conteneur Podman, derrière nginx en reverse proxy sur
`comments.neurau.eu`.

### 1. DNS

Chez le registrar, ajouter un enregistrement `A` (et `AAAA` si IPv6)
pointant `comments.neurau.eu` vers l'IP du VPS.

### 2. Installation sur le VPS

Cloner le repo sur le VPS puis lancer l'installation :

    sudo make vps-install

Cela installe tout : le conteneur Isso (Podman/Quadlet), nginx en
reverse proxy, certbot pour le certificat TLS, et les mises à jour
automatiques du conteneur et du certificat.

Tous les fichiers de configuration dans `/etc` et `/srv` sont des
liens symboliques vers le repo. Pour appliquer une modification,
il suffit de `git pull` puis `sudo make vps-nginx` ou
`sudo make vps-isso` selon ce qui a changé.

Adapter les paramètres SMTP dans `deploy/isso/isso.cfg` pour la
notification des nouveaux commentaires par email.

### Targets disponibles

    sudo make vps-install   # installation complète
    sudo make vps-isso      # (re)configurer le conteneur Isso
    sudo make vps-nginx     # (re)configurer nginx
    sudo make vps-certbot   # (re)configurer certbot + certificat TLS

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
