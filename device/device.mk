ifndef $(BASE_DIR)
	BASE_DIR := $(shell pwd)
endif

virtualenv: ${BASE_DIR}/venv
	virtualenv --no-site-packages $(BASE_DIR)/venv

device-install-deps: virtualenv ${BASE_DIR}/venv/bin/platformio
	@$(BASE_DIR)/venv/bin/pip install platformio

device-build: device-install-deps
	@$(BASE_DIR)/venv/bin/platformio update
	@$(BASE_DIR)/venv/bin/platformio upgrade
	@$(BASE_DIR)/venv/bin/platformio run -d ${BASE_DIR}
