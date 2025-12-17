build:
	go build

run:
	env $(shell grep -v '^#' .env | xargs) go run . < .secrets

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
