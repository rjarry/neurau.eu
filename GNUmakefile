HUGO ?= hugo

.PHONY: build
build:
	$(HUGO) --minify

.PHONY: serve
serve:
	$(HUGO) server -D

.PHONY: clean
clean:
	rm -rf public/ resources/

# VPS targets (run as root on the VPS)
.PHONY: vps-install
vps-install: vps-isso vps-nginx vps-certbot

.PHONY: vps-isso
vps-isso:
	@mkdir -vp /srv/isso/config /srv/isso/db
	@ln -vsfr deploy/isso/isso.cfg /srv/isso/config/isso.cfg
	@ln -vsfr deploy/isso/isso.container /etc/containers/systemd/isso.container
	systemctl daemon-reload
	systemctl start isso
	systemctl enable --now podman-auto-update.timer

.PHONY: vps-nginx
vps-nginx:
	@ln -vsfr deploy/isso/nginx.conf /etc/nginx/sites-available/comments.neurau.eu
	@ln -vsfr /etc/nginx/sites-available/comments.neurau.eu /etc/nginx/sites-enabled/
	nginx -t
	systemctl reload nginx

.PHONY: vps-certbot
vps-certbot:
	apt-get install -y certbot
	@mkdir -vp /var/www/acme
	@if [ ! -d /etc/letsencrypt/live/comments.neurau.eu ]; then \
		certbot certonly --webroot -w /var/www/acme -d comments.neurau.eu; \
	else \
		echo "certificate already exists"; \
	fi
	systemctl enable --now certbot.timer
