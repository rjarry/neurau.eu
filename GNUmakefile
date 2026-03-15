HUGO ?= hugo
GO ?= go

.PHONY: build
build:
	$(HUGO) --minify

.PHONY: serve
serve:
	$(HUGO) server -D

.PHONY: cms
cms: deploy/cms/neurau-cms
	@:

deploy/cms/neurau-cms: deploy/cms/main.go
	$(GO) build -C deploy/cms -o neurau-cms .

.PHONY: clean
clean:
	rm -rf public/ resources/ neurau-cms

# Deploy targets (run as root on the VPS)
.PHONY: deploy
deploy: deploy-isso deploy-cms deploy-nginx deploy-certbot

deploy/isso/secrets.cfg:
	@echo "error: create deploy/isso/secrets.cfg with SMTP_PASSWORD=xxx ..."
	@fail

envsubst := envsubst '$(addprefix $$SMTP_,HOST PORT TO FROM USERNAME PASSWORD)'

.PHONY: deploy-isso
deploy-isso: deploy/isso/secrets.cfg
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

deploy/cms/env:
	@echo "error: create deploy/cms/env (see deploy/cms/env.example)"
	@false

.PHONY: deploy-cms
deploy-cms: deploy/cms/neurau-cms deploy/cms/env
	install -m 755 deploy/cms/neurau-cms /usr/local/bin/neurau-cms
	install -m 600 deploy/cms/env /etc/neurau-cms.env
	install -m 644 deploy/cms/neurau-cms.service /etc/systemd/system/neurau-cms.service
	mkdir -p /var/www/neurau.eu
	systemctl daemon-reload
	systemctl enable --now neurau-cms

.PHONY: deploy-nginx
deploy-nginx:
	rm -f /etc/nginx/sites-enabled/comments.neurau.eu
	rm -f /etc/nginx/sites-available/comments.neurau.eu
	ln -sfr deploy/cms/nginx.conf /etc/nginx/sites-available/neurau.eu
	ln -sfr /etc/nginx/sites-available/neurau.eu /etc/nginx/sites-enabled/
	nginx -t
	systemctl reload nginx

.PHONY: deploy-certbot
deploy-certbot:
	apt-get install -y certbot
	mkdir -p /var/www/acme
	@if [ ! -d /etc/letsencrypt/live/neurau.eu ]; then \
		certbot certonly --webroot -w /var/www/acme -d neurau.eu -d www.neurau.eu; \
	else \
		echo "neurau.eu: certificate already exists"; \
	fi
	systemctl enable --now certbot.timer
