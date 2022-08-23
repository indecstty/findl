ls output | grep html | grep -oE "[[:digit:]]+" | xargs -L1 -I% wkhtmltopdf output/%.html pdfs/%.pdf
