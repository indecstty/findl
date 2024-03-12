#!/bin/bash

counter=0

for file in output/*.html; do
    ((counter++))
    echo "File $file (Run $counter)"

    # Extract the numeric part from the file name
    number=$(echo "$file" | grep -oE "[[:digit:]]+")

    # Encode special characters in the file name
    encoded_file=$(echo "$file" | sed 's/[^A-Za-z0-9._-]/_/g')

    # Run wkhtmltopdf with the appropriate options
    wkhtmltopdf --enable-local-file-access "$file" "pdfs/$number.pdf"

    # Check if wkhtmltopdf command was successful
    if [ $? -eq 0 ]; then
        echo "DONE"
    else
        echo "Error: wkhtmltopdf command failed for $file"
    fi
done

# Zip the contents of the "pdfs" directory
echo "Zipping the PDF files"
zip -r kukot_$(date '+%d%m%Y').zip -j pdfs/*
