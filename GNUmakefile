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

deploy/isso/secrets.cfg:
	@echo "error: create deploy/isso/secrets.cfg with SMTP_PASSWORD=xxx ..."
	@fail

envsubst := envsubst '$(addprefix $$SMTP_,HOST PORT TO FROM USERNAME PASSWORD)'

.PHONY: isso
isso: deploy/isso/secrets.cfg
	id isso >/dev/null 2>&1 || useradd -r -s /usr/sbin/nologin -d /srv/isso -m isso
	grep -q isso /etc/subuid || usermod --add-subuids 100000-165535 --add-subgids 100000-165535 isso
	loginctl enable-linger isso
	mkdir -p /srv/isso/config /srv/isso/db
	chown -R isso:isso /srv/isso
	set -a && . $< && $(envsubst) < deploy/isso/isso.cfg > /srv/isso/config/isso.cfg
	ln -sfr deploy/isso/isso.container /etc/containers/systemd/isso.container
	systemctl daemon-reload
	systemctl start isso
	systemctl enable --now podman-auto-update.timer

.PHONY: nginx
nginx:
	ln -sfr deploy/isso/nginx.conf /etc/nginx/sites-available/comments.neurau.eu
	ln -sfr /etc/nginx/sites-available/comments.neurau.eu /etc/nginx/sites-enabled/
	nginx -t
	systemctl reload nginx

.PHONY: certbot
certbot:
	apt-get install -y certbot
	mkdir -p /var/www/acme
	@if [ ! -d /etc/letsencrypt/live/comments.neurau.eu ]; then \
		certbot certonly --webroot -w /var/www/acme -d comments.neurau.eu; \
	else \
		echo "certificate already exists"; \
	fi
	systemctl enable --now certbot.timer
