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
	sed -i '' -E -e "s/VERSION=.{0,}$$/VERSION=$(shell git log --format=format:'%H' | head -n 1)/" ./.env

build: update-dependencies env-set-version minify-js
	go build

run:
	env $(shell grep -v '^#' .env | xargs) ./wholth_go < .secrets

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
