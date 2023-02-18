GO_NETWORKS = ../minigame-go-networks
DOCKER_YML = build/package/docker-compose.yaml
DOCKER_COMPOSE_COMMAND = docker-compose -f $(DOCKER_YML) --project-directory ./

help: 
	@echo "Help:"
	@echo 	"make Help - show help"
	@echo 	"make up - docker compose up"
	@echo 	"make clean - docker compose down"
	@echo 	"make stop - docker compose stop"
	@echo 	"make rebuild - rebuild containers"
	@echo 	"make restart - restart containers
	@echo 	"make hotreload_on - enable hot reloading; make up or make build must be called if needed to apply changes"
	@echo 	"make hotreload_off - disable hot reloading"
	@echo 	"make network_up - network docker compose up"
	@echo 	"make network_clean - network docker compose down"
	@echo 	"make network_restart - network restart containers"
	@echo 	"make network_db_dump - network db dump"
	@echo 	"make network_db_restore - network db restore"
	
up:
	$(DOCKER_COMPOSE_COMMAND) up -d
	
clean:
	$(DOCKER_COMPOSE_COMMAND) down
	rm -f main

rebuild:
	$(DOCKER_COMPOSE_COMMAND) up -d --build

logs:
	$(DOCKER_COMPOSE_COMMAND) logs -f --tail=10

restart:
	$(DOCKER_COMPOSE_COMMAND) restart

restart_api:
	$(DOCKER_COMPOSE_COMMAND) restart api

restart_ws_loltower:
	$(DOCKER_COMPOSE_COMMAND) restart ws_loltower

restart_ws_lolcouple:
	$(DOCKER_COMPOSE_COMMAND) restart ws_lolcouple

restart_gl_loltower:
	$(DOCKER_COMPOSE_COMMAND) restart gl_loltower

restart_gl_lolcouple:
	$(DOCKER_COMPOSE_COMMAND) restart gl_lolcouple

restart_nginx:
	$(DOCKER_COMPOSE_COMMAND) restart nginx

stop_all:
	$(DOCKER_COMPOSE_COMMAND) stop

stop:
	$(DOCKER_COMPOSE_COMMAND) stop api ws_loltower ws_lolcouple gl_loltower gl_lolcouple

hotreload_on:
	touch .env
	echo 'DOCKERFILE_BASE="local.hotreload"' >> .env

hotreload_off:
	rm -f .env

network_up:
	make -C $(GO_NETWORKS) up

network_clean:
	make -C $(GO_NETWORKS) clean

network_restart:
	make -C $(GO_NETWORKS) restart

network_db_dump:
	make -C $(GO_NETWORKS) up
	make -C $(GO_NETWORKS) db_dump

network_db_restore:
	make -C $(GO_NETWORKS) up
	make -C $(GO_NETWORKS) db_restore