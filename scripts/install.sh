#!/bin/bash

deno task build:all
chmod +x ../bin/restcli*

mv ../bin/restcli* /usr/local/bin/

echo "Installed !"
