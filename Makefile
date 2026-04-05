update-dependencies:
	git submodule update --remote
	cd wholth_lib; make prep-release
	cd wholth_lib; make build-release
	cd wholth_lib; make install-release

minify-js:
	sed -E "/(\/\/)|(\/\*)|(\*\/)/d" static/main.js \
		| awk '{sub(/^ +/,"")}1' \
		| tr -d '\n' \
		> static/main.min.js

env-set-version:
ifeq (,$(findstring darwin, $(shell uname -s)))
	sed -i '' -E -e "s/VERSION=.{0,}$$/VERSION=$(shell git log --format=format:'%H' | head -n 1)/" .env
else
	sed -i -E -e "s/VERSION=.{0,}$$/VERSION=$(shell git log --format=format:'%H' | head -n 1)/" .env
endif

go-build:
	go build

build: update-dependencies env-set-version minify-js go-build

run:
	@-rm .secrets.tmp
	@echo "> SECRETS.DECRYPT.START"
	@gpg --decrypt --output .secrets.tmp .secrets.gpg
	@echo "> SECRETS.DECRYPT.END"
	@env $(shell grep -v '^#' .env | xargs) go run . < .secrets.tmp & rm -f .secrets.tmp ; wait

css-palette:
	sass -s compressed static/palette.scss static/palette.css

css-palette-watch:
	sass -w -s compressed static/palette.scss static/palette.css

css:
	sass -s compressed static/styles.scss static/styles.css

css-watch:
	sass -w -s compressed static/styles.scss static/styles.css

# minify-html-templates:
# 	find templates -name "*.html" | xargs -I{} sh -c 'tr "\n" " " < {} | sed -E "s/ {1,}/ /g" > {}.min'

# cp ~/Projects/wholth/build/libwholth.dylib /usr/local/lib/.
