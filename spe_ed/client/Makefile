all: serve

setup:
	yarn install

clean:
	rm -rf dist

reset:
	rm -rf node_modules .cache

serve:
	npx parcel index.html

dist: clean
	npx parcel build index.html --public-url /spe_ed --no-source-maps
