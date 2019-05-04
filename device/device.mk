ifndef $(BASE_DIR)
	BASE_DIR := $(shell pwd)
endif

ifndef ${OS}
	OS := $(shell uname -s)
endif

HAS_VIRTUALENV := $(shell which virutalenv > /dev/null && echo $?)
HAS_PIP := $(shell which pip > /dev/null && echo $?)

BOLD := "\\e[1m"
GREEN := "\\e[32m"
RESET := "\\e[0m"

install-pip:
ifeq (${HAS_PIP},0)
	@echo "${BOLD}${GREEN}pip is installed${RESET}"
else ifeq (${OS},Linux)
	@echo "${BOLD}${GREEN}installing pip${RESET}"
	@sudo apt install python-pip
	@echo "${BOLD}${GREEN}done${RESET}"
else
	@echo "${BOLD}${GREEN}installing pip${RESET}"
	@brew install python-pip
	@echo "${BOLD}${GREEN}done${RESET}"
endif
	
install-virtualenv: install-pip
ifneq (${HAS_VIRTUALENV},0)
	@echo "${BOLD}${GREEN}installing virtualenv${RESET}"
	@sudo pip install virtualenv
	@echo "${BOLD}${GREEN}done${RESET}"
endif

virtualenv: install-virtualenv
	@echo "${BOLD}${GREEN}creating virtualenv @${RESET}${BOLD}$(BASE_DIR)/venv${RESET}"
	virtualenv --no-site-packages $(BASE_DIR)/venv
	@echo "${BOLD}${GREEN}done${RESET}"

device-install-deps: virtualenv 
	@echo "${BOLD}${GREEN}installing platformio in virtualenv${RESET}"
	@$(BASE_DIR)/venv/bin/pip install platformio
	@echo "${BOLD}${GREEN}done${RESET}"

device-build: device-install-deps
	@$(BASE_DIR)/venv/bin/platformio update
	@$(BASE_DIR)/venv/bin/platformio upgrade
	@$(BASE_DIR)/venv/bin/platformio run -d ${BASE_DIR}
